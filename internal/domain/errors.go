package domain

import "errors"

var (
	ErrArchNotSupported       = errors.New("architecture not supported")
	ErrVariantNotSupported    = errors.New("variant not supported")
	ErrModuleNotFound         = errors.New("module not found")
	ErrModuleAlreadyInstalled = errors.New("module already installed")
	ErrMissingDependency      = errors.New("missing required module dependency")
	ErrIncompatibleModule     = errors.New("module incompatible with project architecture or variant")
	ErrProjectNotFound        = errors.New("archforge.yaml not found in directory")
	ErrInvalidConfig          = errors.New("invalid archforge.yaml configuration")
	ErrUpdateCheckFailed      = errors.New("failed to check for updates")
	ErrAlreadyUpToDate        = errors.New("arch_forge is already up to date")
	ErrDevVersion             = errors.New("cannot auto-update a dev build; install a tagged release first")
	ErrDomainAlreadyExists    = errors.New("a domain with that name already exists")
)
