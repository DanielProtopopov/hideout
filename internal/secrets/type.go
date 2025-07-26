package secrets

import (
	"context"
	"hideout/internal/common/generics"
)

type (
	Secret struct {
		ID     uint   `json:"ID" db:"id" gorm:"column:id;primaryKey;autoIncrement" description:"Primary unique identifier" example:"1"`
		PathID uint   `json:"PathID" db:"path_id" gorm:"column:parent_id" description:"Path unique identifier (link)" example:"0"`
		UID    string `json:"UID" db:"uid" gorm:"column:uid;unique" description:"Secondary unique identifier" example:"abc-def-ghi"`
		Name   string `json:"Name" db:"name" gorm:"column:name" description:"Secret name" example:"DEBUG"`
		Value  string `json:"Value" db:"value" gorm:"column:value" description:"Secret value" example:"Test"`
		Type   string `json:"Type" db:"type" gorm:"column:type" description:"Secret value type" example:"int"`
	}

	Repository interface {
		Get(ctx context.Context, params ListSecretParams) ([]*Secret, error)
		GetMapByID(ctx context.Context, params ListSecretParams) (map[uint]*Secret, error)
		GetMapByUID(ctx context.Context, params ListSecretParams) (map[string]*Secret, error)
		GetMapByPath(ctx context.Context, params ListSecretParams) (map[uint][]*Secret, error)
		GetByUID(ctx context.Context, uid string) (*Secret, error)
		GetByID(ctx context.Context, id uint) (*Secret, error)
		Update(ctx context.Context, id uint, value string) (*Secret, error)
		Create(ctx context.Context, pathID uint, name string, value string, valueType string) (*Secret, error)
		// Delete(ctx context.Context, id uint) error
	}

	ListSecretParams struct {
		generics.ListParams
		PathIDs []uint
		Name    string
		Types   []string
	}
)
