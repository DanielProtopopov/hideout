package model

import (
	"database/sql"
	"time"
)

type Model struct {
	ID        uint         `db:"id" gorm:"column:id;primaryKey;autoIncrement"`
	CreatedAt time.Time    `db:"created_at" gorm:"column:created_at"`
	UpdatedAt time.Time    `db:"updated_at" gorm:"column:updated_at"`
	DeletedAt sql.NullTime `db:"deleted_at" gorm:"column:deleted_at"`
}
