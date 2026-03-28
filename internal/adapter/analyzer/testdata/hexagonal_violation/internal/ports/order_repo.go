package ports

// OrderRepository defines the port for order persistence.
type OrderRepository interface {
	Save(id string) error
}
