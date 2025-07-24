package paths

import (
	"context"
	"hideout/internal/common/generics"
)

type (
	Path struct {
		ID       uint   `json:"ID" db:"id" gorm:"column:id;primaryKey;autoIncrement" description:"Primary unique identifier" example:"1"`
		ParentID uint   `json:"ParentID" db:"parent_id" gorm:"column:parent_id" description:"Parent value identifier (link)" example:"0"`
		UID      string `json:"UID" db:"uid" gorm:"column:uid;unique" description:"Secondary unique identifier" example:"abc-def-ghi"`
		Name     string `json:"Name" db:"name" gorm:"column:name" description:"Folder path name" example:"/"`
	}

	Repository interface {
		Get(ctx context.Context, params ListPathParams) ([]*Path, error)
		GetMapByID(ctx context.Context, params ListPathParams) (map[uint]*Path, error)
		GetMapByUID(ctx context.Context, params ListPathParams) (map[string]*Path, error)
		GetByUID(ctx context.Context, uid string) (*Path, error)
		GetByID(ctx context.Context, id uint) (*Path, error)
		Update(ctx context.Context, id uint, value string) (*Path, error)
		Create(ctx context.Context, pathID uint, name string) (*Path, error)
		Delete(ctx context.Context, id uint) error
	}

	ListPathParams struct {
		generics.ListParams
		Name string
	}
)
