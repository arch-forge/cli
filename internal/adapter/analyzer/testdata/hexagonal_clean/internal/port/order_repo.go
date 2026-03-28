package port

import _ "example.com/cleanapp/internal/domain"

// OrderRepository defines the port for order persistence.
// Ports may import domain — this is allowed.
type OrderRepository interface {
	Save(id string) error
}
