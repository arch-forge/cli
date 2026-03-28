// Package updater provides a GitHub-backed implementation of port.Updater.
package updater

import (
	"archive/tar"
	"archive/zip"
	"bufio"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/arch-forge/cli/internal/domain"
	"github.com/arch-forge/cli/internal/port"
)

const (
	// githubAPIURL is the live GitHub releases endpoint.
	githubAPIURL = "https://api.github.com/repos/archforge/cli/releases/latest"
	httpTimeout  = 30 * time.Second
)

// githubRelease is the subset of the GitHub releases API response that
// GithubUpdater needs.
type githubRelease struct {
	TagName string        `json:"tag_name"`
	Assets  []githubAsset `json:"assets"`
}

// githubAsset represents a single file attached to a GitHub release.
type githubAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
}

// GithubUpdater implements port.Updater by querying the GitHub releases API.
type GithubUpdater struct {
	client *http.Client
	apiURL string
}

// NewGithubUpdater returns a GithubUpdater with a sensible default HTTP client
// pointing at the live GitHub releases endpoint.
func NewGithubUpdater() *GithubUpdater {
	return &GithubUpdater{
		client: &http.Client{Timeout: httpTimeout},
		apiURL: githubAPIURL,
	}
}

// NewGithubUpdaterWithURL returns a GithubUpdater that uses apiURL as the
// releases endpoint. Intended for testing with httptest servers.
func NewGithubUpdaterWithURL(apiURL string) *GithubUpdater {
	return &GithubUpdater{
		client: &http.Client{Timeout: httpTimeout},
		apiURL: apiURL,
	}
}

// LatestRelease queries the GitHub releases API and returns metadata about the
// latest available release for the current OS and architecture.
// Returns domain.ErrUpdateCheckFailed on any network or parsing failure.
func (g *GithubUpdater) LatestRelease() (port.ReleaseInfo, error) {
	req, err := http.NewRequest(http.MethodGet, g.apiURL, nil)
	if err != nil {
		return port.ReleaseInfo{}, fmt.Errorf("update: build request: %w", domain.ErrUpdateCheckFailed)
	}

	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "arch_forge")

	resp, err := g.client.Do(req)
	if err != nil {
		return port.ReleaseInfo{}, fmt.Errorf("update: fetch latest release: %w", domain.ErrUpdateCheckFailed)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return port.ReleaseInfo{}, fmt.Errorf("update: GitHub API returned status %d: %w", resp.StatusCode, domain.ErrUpdateCheckFailed)
	}

	var release githubRelease
	if err = json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return port.ReleaseInfo{}, fmt.Errorf("update: decode release JSON: %w", domain.ErrUpdateCheckFailed)
	}

	binaryName := assetName(release.TagName)

	binaryAsset, found := findAsset(release.Assets, binaryName)
	if !found {
		return port.ReleaseInfo{}, fmt.Errorf("update: no binary asset for %s/%s: %w", runtime.GOOS, runtime.GOARCH, domain.ErrUpdateCheckFailed)
	}

	checksum := g.fetchChecksum(release.Assets, binaryName)

	return port.ReleaseInfo{
		TagName:     release.TagName,
		DownloadURL: binaryAsset.BrowserDownloadURL,
		Checksum:    checksum,
	}, nil
}

// DownloadBinary fetches the archive at info.DownloadURL, extracts the
// arch_forge binary, optionally verifies its SHA-256 checksum, and writes
// the result to destPath with mode 0755.
// Returns domain.ErrUpdateCheckFailed on any failure.
func (g *GithubUpdater) DownloadBinary(info port.ReleaseInfo, destPath string) error {
	resp, err := g.client.Get(info.DownloadURL) //nolint:noctx
	if err != nil {
		return fmt.Errorf("update: download binary archive: %w", domain.ErrUpdateCheckFailed)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("update: download returned status %d: %w", resp.StatusCode, domain.ErrUpdateCheckFailed)
	}

	if strings.HasSuffix(info.DownloadURL, ".zip") {
		if err = extractZip(resp.Body, destPath); err != nil {
			return fmt.Errorf("update: extract zip: %w", domain.ErrUpdateCheckFailed)
		}
	} else {
		if err = extractTarGz(resp.Body, destPath); err != nil {
			return fmt.Errorf("update: extract tar.gz: %w", domain.ErrUpdateCheckFailed)
		}
	}

	if err = os.Chmod(destPath, 0755); err != nil {
		return fmt.Errorf("update: set permissions: %w", domain.ErrUpdateCheckFailed)
	}

	if info.Checksum != "" {
		if err = verifyChecksum(destPath, info.Checksum); err != nil {
			return err
		}
	}

	return nil
}

// assetName builds the GoReleaser archive filename for the current OS/arch.
//
// GoReleaser naming convention:
//
//	arch_forge_{version}_{os}_{arch}.tar.gz   (linux, darwin)
//	arch_forge_{version}_{os}_{arch}.zip      (windows)
//
// The version component strips the leading "v" from the tag (e.g. "v1.2.0" → "1.2.0").
func assetName(tagName string) string {
	version := strings.TrimPrefix(tagName, "v")
	ext := ".tar.gz"
	if runtime.GOOS == "windows" {
		ext = ".zip"
	}
	return fmt.Sprintf("arch_forge_%s_%s_%s%s", version, runtime.GOOS, runtime.GOARCH, ext)
}

// findAsset searches assets for an entry whose Name matches name exactly.
func findAsset(assets []githubAsset, name string) (githubAsset, bool) {
	for _, a := range assets {
		if a.Name == name {
			return a, true
		}
	}
	return githubAsset{}, false
}

// fetchChecksum downloads the checksums.txt asset (if present) and extracts
// the SHA-256 hex digest for binaryName. Returns an empty string on any
// failure so that the caller can treat a missing checksum as non-fatal.
func (g *GithubUpdater) fetchChecksum(assets []githubAsset, binaryName string) string {
	csAsset, found := findAsset(assets, "checksums.txt")
	if !found {
		return ""
	}

	resp, err := g.client.Get(csAsset.BrowserDownloadURL) //nolint:noctx
	if err != nil {
		return ""
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return ""
	}

	return parseChecksum(resp.Body, binaryName)
}

// parseChecksum reads a GoReleaser checksums.txt and returns the hex digest
// for the given filename. GoReleaser v2 format per line:
//
//	<sha256hex>  <filename>
//
// (two spaces between digest and filename, no "sha256:" prefix).
func parseChecksum(r io.Reader, filename string) string {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := scanner.Text()
		// Split on double-space as specified by GoReleaser v2.
		parts := strings.SplitN(line, "  ", 2)
		if len(parts) != 2 {
			continue
		}
		if strings.TrimSpace(parts[1]) == filename {
			return strings.TrimSpace(parts[0])
		}
	}
	return ""
}

// extractTarGz reads a gzip-compressed tar stream from r, finds the entry
// whose base name is "arch_forge", and writes it to destPath.
func extractTarGz(r io.Reader, destPath string) error {
	gzr, err := gzip.NewReader(r)
	if err != nil {
		return fmt.Errorf("open gzip reader: %w", err)
	}
	defer gzr.Close()

	tr := tar.NewReader(gzr)

	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("read tar entry: %w", err)
		}

		if filepath.Base(hdr.Name) != "arch_forge" {
			continue
		}

		if err = writeFile(tr, destPath); err != nil {
			return err
		}
		return nil
	}

	return fmt.Errorf("arch_forge binary not found in archive")
}

// extractZip buffers the zip archive from r into a temporary file (zip
// requires a seekable reader), then finds "arch_forge.exe" and writes it to
// destPath.
func extractZip(r io.Reader, destPath string) error {
	tmp, err := os.CreateTemp("", "arch_forge_update_*.zip")
	if err != nil {
		return fmt.Errorf("create temp zip file: %w", err)
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)

	if _, err = io.Copy(tmp, r); err != nil {
		tmp.Close()
		return fmt.Errorf("buffer zip to disk: %w", err)
	}
	tmp.Close()

	zr, err := zip.OpenReader(tmpPath)
	if err != nil {
		return fmt.Errorf("open zip reader: %w", err)
	}
	defer zr.Close()

	for _, f := range zr.File {
		if filepath.Base(f.Name) != "arch_forge.exe" {
			continue
		}

		rc, err := f.Open()
		if err != nil {
			return fmt.Errorf("open zip entry: %w", err)
		}

		writeErr := writeFile(rc, destPath)
		rc.Close()
		if writeErr != nil {
			return writeErr
		}
		return nil
	}

	return fmt.Errorf("arch_forge.exe not found in archive")
}

// writeFile copies src to a new file at destPath, truncating it first if it
// already exists.
func writeFile(src io.Reader, destPath string) error {
	out, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return fmt.Errorf("open destination file: %w", err)
	}
	defer out.Close()

	if _, err = io.Copy(out, src); err != nil {
		return fmt.Errorf("write destination file: %w", err)
	}
	return nil
}

// verifyChecksum computes the SHA-256 digest of the file at path and compares
// it against the expected hex string. Returns domain.ErrUpdateCheckFailed on
// mismatch.
func verifyChecksum(path, expected string) error {
	f, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("update: open file for checksum: %w", domain.ErrUpdateCheckFailed)
	}
	defer f.Close()

	h := sha256.New()
	if _, err = io.Copy(h, f); err != nil {
		return fmt.Errorf("update: compute checksum: %w", domain.ErrUpdateCheckFailed)
	}

	actual := hex.EncodeToString(h.Sum(nil))
	if actual != expected {
		return fmt.Errorf("update: checksum mismatch: %w", domain.ErrUpdateCheckFailed)
	}
	return nil
}
