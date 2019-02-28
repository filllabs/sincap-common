package dbutil

import "time"

type Model struct {
	ID        uint `gorm:"primary_key"`
	CreatedAt *time.Time
	UpdatedAt *time.Time
	DeletedAt *time.Time `gorm:"index"`
}

// PolymorphicModel helps to make your model having a polymorphic relation.
type PolymorphicModel struct {
	HolderID   uint   `gorm:"index;not null"`
	HolderType string `gorm:"index;not null"`
}

// OwnedModel helps to make your model having a owner restriction
type OwnedModel struct {
	OwnerID uint `gorm:"index;not null"`
}
