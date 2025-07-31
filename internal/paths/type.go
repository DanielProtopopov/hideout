package paths

import (
	"context"
	"hideout/internal/common/generics"
	"hideout/internal/common/model"
)

type (
	Path struct {
		model.Model
		ParentID uint   `json:"ParentID" db:"parent_id" gorm:"column:parent_id" description:"Parent value identifier (link)" example:"0"`
		UID      string `json:"UID" db:"uid" gorm:"column:uid;unique" description:"Secondary unique identifier" example:"abc-def-ghi"`
		Name     string `json:"Name" db:"name" gorm:"column:name" description:"Folder path name" example:"/"`
	}

	Repository interface {
		GetID(ctx context.Context) (uint, error)
		Get(ctx context.Context, params ListPathParams) ([]*Path, error)
		GetMapByID(ctx context.Context, params ListPathParams) (map[uint]*Path, error)
		GetMapByUID(ctx context.Context, params ListPathParams) (map[string]*Path, error)
		GetByUID(ctx context.Context, uid string) (*Path, error)
		GetByID(ctx context.Context, id uint) (*Path, error)
		Update(ctx context.Context, id uint, value string) (*Path, error)
		Create(ctx context.Context, id uint, uid string, parentPathID uint, name string) (*Path, error)
		Delete(ctx context.Context, id uint, forceDelete bool) error
	}

	ListPathParams struct {
		generics.ListParams
		Name         string
		ParentPathID uint
	}

	// multiSorter implements the Sort interface, sorting the secrets within.
	multiSorter struct {
		paths []*Path
		less  []lessFunc
	}
)
