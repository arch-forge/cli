package updater_test

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/arch-forge/cli/internal/adapter/updater"
	"github.com/arch-forge/cli/internal/domain"
	"github.com/arch-forge/cli/internal/port"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// assetNameForTest mirrors the naming logic inside github_updater.go so the
// tests can build a matching asset name without accessing the unexported helper.
func assetNameForTest(tagName string) string {
	version := strings.TrimPrefix(tagName, "v")
	ext := ".tar.gz"
	if runtime.GOOS == "windows" {
		ext = ".zip"
	}
	return fmt.Sprintf("arch_forge_%s_%s_%s%s", version, runtime.GOOS, runtime.GOARCH, ext)
}

// githubReleasePayload is a minimal GitHub releases API response used by tests.
type githubReleasePayload struct {
	TagName string `json:"tag_name"`
	Assets  []struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	} `json:"assets"`
}

// newTestUpdater constructs a GithubUpdater whose HTTP client points at the
// provided test server instead of api.github.com.
func newTestUpdater(apiURL string) *updater.GithubUpdater {
	return updater.NewGithubUpdaterWithURL(apiURL)
}

// buildTarGz returns an in-memory .tar.gz archive containing a single entry
// named "arch_forge" with the given content.
func buildTarGz(content []byte) []byte {
	var buf bytes.Buffer
	gzw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gzw)

	hdr := &tar.Header{
		Name: "arch_forge",
		Mode: 0755,
		Size: int64(len(content)),
	}
	_ = tw.WriteHeader(hdr)
	_, _ = tw.Write(content)
	_ = tw.Close()
	_ = gzw.Close()

	return buf.Bytes()
}

func TestGithubUpdater_LatestRelease_Success(t *testing.T) {
	const tagName = "v1.5.0"
	assetName := assetNameForTest(tagName)
	downloadURL := "https://example.com/" + assetName

	payload := githubReleasePayload{TagName: tagName}
	payload.Assets = append(payload.Assets, struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	}{Name: assetName, BrowserDownloadURL: downloadURL})

	body, err := json.Marshal(payload)
	require.NoError(t, err)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	u := newTestUpdater(srv.URL)
	info, err := u.LatestRelease()

	require.NoError(t, err)
	assert.Equal(t, tagName, info.TagName)
	assert.Equal(t, downloadURL, info.DownloadURL)
}

func TestGithubUpdater_LatestRelease_NoMatchingAsset(t *testing.T) {
	const tagName = "v1.5.0"

	// Provide an asset for a different OS/arch so the matcher fails.
	payload := githubReleasePayload{TagName: tagName}
	payload.Assets = append(payload.Assets, struct {
		Name               string `json:"name"`
		BrowserDownloadURL string `json:"browser_download_url"`
	}{Name: "arch_forge_1.5.0_plan9_mips.tar.gz", BrowserDownloadURL: "https://example.com/wrong"})

	body, err := json.Marshal(payload)
	require.NoError(t, err)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	u := newTestUpdater(srv.URL)
	_, err = u.LatestRelease()

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrUpdateCheckFailed)
}

func TestGithubUpdater_LatestRelease_HTTPError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	u := newTestUpdater(srv.URL)
	_, err := u.LatestRelease()

	require.Error(t, err)
	assert.ErrorIs(t, err, domain.ErrUpdateCheckFailed)
}

func TestGithubUpdater_DownloadBinary_TarGz(t *testing.T) {
	fakeBinaryContent := []byte("#!/bin/sh\necho hello")
	archive := buildTarGz(fakeBinaryContent)

	const tagName = "v1.5.0"
	assetName := assetNameForTest(tagName)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(archive)
	}))
	defer srv.Close()

	u := newTestUpdater(srv.URL)

	destPath := filepath.Join(t.TempDir(), "arch_forge_downloaded")
	info := port.ReleaseInfo{
		TagName:     tagName,
		DownloadURL: srv.URL + "/" + assetName, // ends with .tar.gz on non-Windows
	}

	// Windows uses .zip; skip the tar.gz variant on Windows.
	if runtime.GOOS == "windows" {
		t.Skip("tar.gz test skipped on Windows")
	}

	err := u.DownloadBinary(info, destPath)
	require.NoError(t, err)

	// Verify the file was written with the correct content.
	got, err := os.ReadFile(destPath)
	require.NoError(t, err)
	assert.Equal(t, fakeBinaryContent, got)

	// Verify the file has executable permissions.
	fi, err := os.Stat(destPath)
	require.NoError(t, err)
	assert.True(t, fi.Mode()&0100 != 0, "binary should be executable")
}
