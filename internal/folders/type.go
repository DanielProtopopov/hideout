package folders

import (
	"context"
	"hideout/internal/common/generics"
	"hideout/internal/common/model"
)

type (
	Folder struct {
		model.Model
		ParentID uint   `json:"ParentID" db:"parent_id" gorm:"column:parent_id" description:"Parent value identifier (link)" example:"0"`
		UID      string `json:"UID" db:"uid" gorm:"column:uid;unique" description:"Secondary unique identifier" example:"abc-def-ghi"`
		Name     string `json:"Name" db:"name" gorm:"column:name" description:"Folder folder name" example:"/"`
	}

	Repository interface {
		GetID(ctx context.Context) (uint, error)
		Get(ctx context.Context, params ListFolderParams) ([]*Folder, error)
		GetMapByID(ctx context.Context, params ListFolderParams) (map[uint]*Folder, error)
		GetMapByUID(ctx context.Context, params ListFolderParams) (map[string]*Folder, error)
		GetByUID(ctx context.Context, uid string) (*Folder, error)
		GetByID(ctx context.Context, id uint) (*Folder, error)
		Update(ctx context.Context, folder Folder) (*Folder, error)
		Create(ctx context.Context, folder Folder) (*Folder, error)
		Delete(ctx context.Context, id uint, forceDelete bool) error
		Count(ctx context.Context, params ListFolderParams) (uint, error)
		Load(ctx context.Context) ([]Folder, error)
	}

	ListFolderParams struct {
		generics.ListParams
		Name           string
		ParentFolderID uint
	}

	// multiSorter implements the Sort interface, sorting the secrets within.
	multiSorter struct {
		folders []*Folder
		less    []lessFunc
	}
)
