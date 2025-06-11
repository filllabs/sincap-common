package interfaces

// TableNamer interface for models that can provide their table name
type TableNamer interface {
	TableName() string
}

// IDGetter interface for models that can provide their ID
type IDGetter interface {
	GetID() interface{}
}

// IDSetter interface for models that can set their ID
type IDSetter interface {
	SetID(id interface{}) error
}

// FieldMapper interface for models that can provide their field mappings
type FieldMapper interface {
	GetFieldMap() map[string]interface{}
}
