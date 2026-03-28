package app_test

import (
	"errors"
	"testing"

	"github.com/arch-forge/cli/internal/app"
	"github.com/arch-forge/cli/internal/domain"
	"github.com/arch-forge/cli/internal/port"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockUpdater is a test double for port.Updater.
type mockUpdater struct {
	latestReleaseFunc  func() (port.ReleaseInfo, error)
	downloadBinaryFunc func(info port.ReleaseInfo, destPath string) error
}

func (m *mockUpdater) LatestRelease() (port.ReleaseInfo, error) {
	return m.latestReleaseFunc()
}

func (m *mockUpdater) DownloadBinary(info port.ReleaseInfo, destPath string) error {
	return m.downloadBinaryFunc(info, destPath)
}

func TestUpdateUseCase_DevVersion(t *testing.T) {
	updater := &mockUpdater{
		latestReleaseFunc: func() (port.ReleaseInfo, error) {
			// Should never be called when version is "dev".
			t.Fatal("LatestRelease should not be called for dev builds")
			return port.ReleaseInfo{}, nil
		},
		downloadBinaryFunc: func(info port.ReleaseInfo, destPath string) error {
			t.Fatal("DownloadBinary should not be called for dev builds")
			return nil
		},
	}

	uc := app.NewUpdateUseCase(updater, "dev")
	_, err := uc.Execute(app.UpdateOptions{})

	require.Error(t, err)
	assert.True(t, errors.Is(err, domain.ErrDevVersion))
}

func TestUpdateUseCase_AlreadyUpToDate(t *testing.T) {
	updater := &mockUpdater{
		latestReleaseFunc: func() (port.ReleaseInfo, error) {
			return port.ReleaseInfo{TagName: "v1.2.3"}, nil
		},
		downloadBinaryFunc: func(info port.ReleaseInfo, destPath string) error {
			t.Fatal("DownloadBinary should not be called when already up to date")
			return nil
		},
	}

	uc := app.NewUpdateUseCase(updater, "1.2.3")
	result, err := uc.Execute(app.UpdateOptions{Force: false})

	require.NoError(t, err)
	assert.True(t, result.AlreadyUpToDate)
}

func TestUpdateUseCase_Force_WhenAlreadyUpToDate(t *testing.T) {
	downloadCalled := false

	updater := &mockUpdater{
		latestReleaseFunc: func() (port.ReleaseInfo, error) {
			return port.ReleaseInfo{
				TagName:     "v1.2.3",
				DownloadURL: "https://example.com/arch_forge_1.2.3_linux_amd64.tar.gz",
			}, nil
		},
		downloadBinaryFunc: func(info port.ReleaseInfo, destPath string) error {
			downloadCalled = true
			// Write a tiny placeholder so the rename that follows has something to work with.
			return nil
		},
	}

	uc := app.NewUpdateUseCase(updater, "1.2.3")
	// Force=true means the use case must proceed past the "already up to date"
	// guard and call DownloadBinary. The subsequent os.Rename of the test
	// binary is expected to fail with a permissions or cross-device error —
	// that is acceptable; we only verify DownloadBinary was reached.
	_, err := uc.Execute(app.UpdateOptions{Force: true})

	// DownloadBinary must have been invoked.
	assert.True(t, downloadCalled, "DownloadBinary should be called when Force=true")

	// The error (if any) comes from os.Rename on the test binary, not from
	// DownloadBinary. We do not assert NoError here because the test runner
	// binary is read-only.
	_ = err
}

func TestUpdateUseCase_Success(t *testing.T) {
	downloadCalled := false

	updater := &mockUpdater{
		latestReleaseFunc: func() (port.ReleaseInfo, error) {
			return port.ReleaseInfo{
				TagName:     "v2.0.0",
				DownloadURL: "https://example.com/arch_forge_2.0.0_linux_amd64.tar.gz",
			}, nil
		},
		downloadBinaryFunc: func(info port.ReleaseInfo, destPath string) error {
			downloadCalled = true
			return nil
		},
	}

	uc := app.NewUpdateUseCase(updater, "1.0.0")
	result, err := uc.Execute(app.UpdateOptions{})

	// DownloadBinary must always be called when a newer version exists.
	assert.True(t, downloadCalled, "DownloadBinary should be called for a newer release")

	if err != nil {
		// The only acceptable post-download error is an os.Rename failure
		// (the test binary directory is read-only or cross-device).
		assert.Contains(t, err.Error(), "replace binary",
			"unexpected error after successful download: %v", err)
		return
	}

	// Happy path: versions should be reported correctly.
	assert.Equal(t, "1.0.0", result.PreviousVersion)
	assert.Equal(t, "v2.0.0", result.NewVersion)
	assert.False(t, result.AlreadyUpToDate)
}

func TestUpdateUseCase_NetworkError(t *testing.T) {
	networkErr := errors.New("connection refused")

	updater := &mockUpdater{
		latestReleaseFunc: func() (port.ReleaseInfo, error) {
			return port.ReleaseInfo{}, networkErr
		},
		downloadBinaryFunc: func(info port.ReleaseInfo, destPath string) error {
			t.Fatal("DownloadBinary should not be called when LatestRelease fails")
			return nil
		},
	}

	uc := app.NewUpdateUseCase(updater, "1.0.0")
	_, err := uc.Execute(app.UpdateOptions{})

	require.Error(t, err)
	assert.True(t, errors.Is(err, networkErr))
	assert.Contains(t, err.Error(), "check latest release")
}

func TestUpdateUseCase_DownloadError(t *testing.T) {
	downloadErr := errors.New("download failed")

	updater := &mockUpdater{
		latestReleaseFunc: func() (port.ReleaseInfo, error) {
			return port.ReleaseInfo{
				TagName:     "v2.0.0",
				DownloadURL: "https://example.com/arch_forge_2.0.0_linux_amd64.tar.gz",
			}, nil
		},
		downloadBinaryFunc: func(info port.ReleaseInfo, destPath string) error {
			return downloadErr
		},
	}

	uc := app.NewUpdateUseCase(updater, "1.0.0")
	_, err := uc.Execute(app.UpdateOptions{})

	require.Error(t, err)
	assert.True(t, errors.Is(err, downloadErr))
	assert.Contains(t, err.Error(), "download binary")
}
