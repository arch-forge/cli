package domain

import (
	// This import violates the hexagonal rule: domain must not import port.
	_ "example.com/testapp/internal/ports"
)

// Order is a sample domain entity.
type Order struct {
	ID string
}
