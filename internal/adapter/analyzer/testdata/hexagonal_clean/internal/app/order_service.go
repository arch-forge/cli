package app

import (
	_ "example.com/cleanapp/internal/domain"
	_ "example.com/cleanapp/internal/port"
)

// OrderService is a sample conformant use case that imports domain and port only.
type OrderService struct{}
