package store

import "gorm.io/gorm"

type store struct {
	db *gorm.DB
}

func New(db *gorm.DB) *store {
	return &store{
		db: db,
	}
}
