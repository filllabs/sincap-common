package dbutil

import "time"

type Model struct {
	ID        uint `gorm:"primary_key"`
	CreatedAt *time.Time
	UpdatedAt *time.Time
	DeletedAt *time.Time `sql:"index"`
}

// PolymorphicModel helps to make your model having a polymorphic relation.
type PolymorphicModel struct {
	HolderID   uint
	HolderType string
}

// OwnedModel helps to make your model having a owner restriction
type OwnedModel struct {
	OwnerID uint
}
