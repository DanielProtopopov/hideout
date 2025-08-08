package model

import (
	"database/sql"
	"time"
)

type Model struct {
	ID        uint         `struc:"uint64" json:"ID" bson:"ID" csv:"ID" xml:"ID" yaml:"ID" db:"id" gorm:"column:id;primaryKey;autoIncrement"`
	CreatedAt time.Time    `json:"CreatedAt" bson:"CreatedAt" csv:"CreatedAt" xml:"CreatedAt" yaml:"CreatedAt" db:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time    `json:"UpdatedAt" bson:"UpdatedAt" csv:"UpdatedAt" xml:"UpdatedAt" yaml:"UpdatedAt" db:"updated_at" gorm:"column:updated_at"`
	DeletedAt sql.NullTime `json:"DeletedAt" bson:"DeletedAt" csv:"DeletedAt" xml:"DeletedAt" yaml:"DeletedAt" db:"deleted_at" gorm:"column:deleted_at"`
}
