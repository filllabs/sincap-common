package dbutil

import (
	"time"

	"gorm.io/gorm"
)

// Model is default model with soft delete
type Model struct {
	ID        uint `gorm:"primary_key"`
	CreatedAt *time.Time
	UpdatedAt *time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

// ModelHD is for hard delete
type ModelHD struct {
	ID        uint `gorm:"primary_key"`
	CreatedAt *time.Time
	UpdatedAt *time.Time
}

// PolymorphicModel helps to make your model having a polymorphic relation.
type PolymorphicModel struct {
	HolderID   uint   `gorm:"index"`
	HolderType string `gorm:"index"`
}

// OwnedModel helps to make your model having a owner restriction
type OwnedModel struct {
	OwnerID uint `gorm:"index;not null"`
}
