package secrets

import (
	"context"
	"hideout/internal/common/generics"
)

type (
	Secret struct {
		ID       uint   `json:"ID" db:"id" gorm:"column:id;primaryKey;autoIncrement" description:"Primary unique identifier" example:"1"`
		ParentID uint   `json:"ParentID" db:"parent_id" gorm:"column:parent_id" description:"Parent value identifier (link)" example:"0"`
		UID      string `json:"UID" db:"uid" gorm:"column:uid;unique" description:"Secondary unique identifier" example:"abc-def-ghi"`
		Path     string `json:"Path" db:"path" gorm:"column:path" description:"Folder path" example:"/"`
		Name     string `json:"Name" db:"name" gorm:"column:name" description:"Secret name" example:"DEBUG"`
		Value    string `json:"Value" db:"value" gorm:"column:value" description:"Secret value" example:"Test"`
		Type     string `json:"Type" db:"type" gorm:"column:type" description:"Secret value type" example:"int"`
	}

	Repository interface {
		NewRepository() *Repository
		GetPaths(ctx context.Context) ([]string, error)
		Get(ctx context.Context, params ListSecretParams) (map[string]*Secret, error)
		GetMap(ctx context.Context, params ListSecretParams) (map[string][]*Secret, error)
		GetMapByPath(ctx context.Context, params ListSecretParams) (map[string][]*Secret, error)
		GetMapByUID(ctx context.Context, params ListSecretParams) (map[string][]*Secret, error)
		GetByUID(ctx context.Context, uid string) (*Secret, error)
		Update(ctx context.Context, uid string, columnsMap map[string]interface{}) (*Secret, error)
		Create(ctx context.Context, columnsMap map[string]interface{}) (*Secret, error)
		Delete(ctx context.Context, uid string) error
	}

	ListSecretParams struct {
		generics.ListParams
		Path  string
		Name  string
		Types []string
	}
)
