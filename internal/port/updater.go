package port

// ReleaseInfo holds metadata about a remote release fetched from the upstream provider.
type ReleaseInfo struct {
	// TagName is the full release tag, e.g. "v1.2.0".
	TagName string
	// DownloadURL is the direct URL to the binary archive for the current OS/arch.
	DownloadURL string
	// Checksum is the SHA-256 hex digest for the archive (empty if unavailable).
	Checksum string
}

// Updater is the port that UpdateUseCase calls to check and fetch new releases.
type Updater interface {
	// LatestRelease queries the upstream release provider and returns metadata
	// about the latest available release. Returns ErrUpdateCheckFailed on
	// network or parsing errors.
	LatestRelease() (ReleaseInfo, error)

	// DownloadBinary fetches the archive at info.DownloadURL, extracts the
	// binary, verifies its SHA-256 checksum against info.Checksum (if
	// non-empty), and writes the executable to destPath with mode 0755.
	// Returns ErrUpdateCheckFailed on network, extraction, or checksum errors.
	DownloadBinary(info ReleaseInfo, destPath string) error
}
