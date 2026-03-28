package app

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/arch-forge/cli/internal/domain"
	"github.com/arch-forge/cli/internal/port"
)

// UpdateOptions carries parameters for the update workflow.
type UpdateOptions struct {
	Force bool // skip "already up to date" check
}

// UpdateResult is returned to the CLI layer for display.
type UpdateResult struct {
	AlreadyUpToDate bool
	PreviousVersion string
	NewVersion      string
}

// UpdateUseCase checks for a new release and replaces the running binary.
type UpdateUseCase struct {
	updater        port.Updater
	currentVersion string
}

// NewUpdateUseCase constructs an UpdateUseCase with the given updater and the
// current binary version as injected by the build toolchain.
func NewUpdateUseCase(updater port.Updater, currentVersion string) *UpdateUseCase {
	return &UpdateUseCase{
		updater:        updater,
		currentVersion: currentVersion,
	}
}

// Execute runs the full update workflow: checks for a newer release, downloads
// the binary, and performs an atomic replacement of the running executable.
func (uc *UpdateUseCase) Execute(opts UpdateOptions) (UpdateResult, error) {
	if uc.currentVersion == "dev" {
		return UpdateResult{}, domain.ErrDevVersion
	}

	info, err := uc.updater.LatestRelease()
	if err != nil {
		return UpdateResult{}, fmt.Errorf("update: check latest release: %w", err)
	}

	if "v"+uc.currentVersion == info.TagName && !opts.Force {
		return UpdateResult{AlreadyUpToDate: true}, nil
	}

	exePath, err := os.Executable()
	if err != nil {
		return UpdateResult{}, fmt.Errorf("update: resolve executable path: %w", err)
	}

	exePath, err = filepath.EvalSymlinks(exePath)
	if err != nil {
		return UpdateResult{}, fmt.Errorf("update: eval symlinks: %w", err)
	}

	tmpFile, err := os.CreateTemp(filepath.Dir(exePath), ".arch_forge_update_*")
	if err != nil {
		return UpdateResult{}, fmt.Errorf("update: create temp file: %w", err)
	}

	tmpPath := tmpFile.Name()
	tmpFile.Close()

	var replaced bool
	defer func() {
		if !replaced {
			os.Remove(tmpPath)
		}
	}()

	if err = uc.updater.DownloadBinary(info, tmpPath); err != nil {
		return UpdateResult{}, fmt.Errorf("update: download binary: %w", err)
	}

	if runtime.GOOS == "windows" {
		oldPath := exePath + ".old"
		os.Remove(oldPath)

		if err = os.Rename(exePath, oldPath); err != nil {
			return UpdateResult{}, fmt.Errorf("update: replace binary: %w", err)
		}

		if err = os.Rename(tmpPath, exePath); err != nil {
			return UpdateResult{}, fmt.Errorf("update: replace binary: %w", err)
		}
	} else {
		if err = os.Rename(tmpPath, exePath); err != nil {
			return UpdateResult{}, fmt.Errorf("update: replace binary: %w", err)
		}
	}

	replaced = true

	return UpdateResult{
		PreviousVersion: uc.currentVersion,
		NewVersion:      info.TagName,
	}, nil
}
