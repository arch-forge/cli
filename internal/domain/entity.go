package domain

// RelationKind represents the type of relationship between entities.
type RelationKind string

const (
	RelBelongsTo RelationKind = "belongs_to"
	RelHasMany   RelationKind = "has_many"
	RelHasOne    RelationKind = "has_one"
)

// Relation describes an association from one entity to another.
type Relation struct {
	Kind   RelationKind
	Target string
}

// Field describes a single field on an entity, including its Go and SQL representations.
type Field struct {
	Name       string
	Type       string
	GoType     string
	SQLType    string
	JSONName   string
	DBName     string
	Nullable   bool
	Validation string
}

// EntityInfo holds the definition of a domain entity used during code generation.
type EntityInfo struct {
	Name      string
	Fields    []Field
	Relations []Relation
}
