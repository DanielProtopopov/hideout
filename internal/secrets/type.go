package secrets

import (
	"context"
	"hideout/internal/common/generics"
	"hideout/internal/common/model"
)

type (
	Secret struct {
		model.Model
		FolderID uint   `json:"FolderID" bson:"FolderID" xml:"FolderID" yaml:"FolderID" csv:"FolderID" db:"folder_id" gorm:"column:folder_id" description:"Folder unique identifier (link)" example:"0"`
		UID      string `json:"UID" bson:"UID" xml:"UID" csv:"UID" yaml:"UID" db:"uid" gorm:"column:uid;unique" description:"Secondary unique identifier" example:"abc-def-ghi"`
		Name     string `json:"Name" bson:"Name" xml:"Name" csv:"Name" yaml:"Name" db:"name" gorm:"column:name" description:"Secret name" example:"DEBUG"`
		Value    string `json:"Value" bson:"Value" xml:"Value" csv:"Value" yaml:"Value" db:"value" gorm:"column:value" description:"Secret value" example:"Test"`
		Script   string `json:"Script" bson:"Script" xml:"Script" csv:"Script" yaml:"Script" db:"script" description:"Script for dynamic value" example:"time.RFC3339"`
	}

	Repository interface {
		GetID(ctx context.Context) (uint, error)
		Get(ctx context.Context, params ListSecretParams) ([]*Secret, error)
		GetMapByID(ctx context.Context, params ListSecretParams) (map[uint]*Secret, error)
		GetMapByUID(ctx context.Context, params ListSecretParams) (map[string]*Secret, error)
		GetMapByFolder(ctx context.Context, params ListSecretParams) (map[uint][]*Secret, error)
		GetByUID(ctx context.Context, uid string) (*Secret, error)
		GetByID(ctx context.Context, id uint) (*Secret, error)
		Update(ctx context.Context, secret Secret) (*Secret, error)
		Create(ctx context.Context, secret Secret) (*Secret, error)
		Delete(ctx context.Context, id uint, forceDelete bool) error
		Count(ctx context.Context, params ListSecretParams) (uint, error)
		Load(ctx context.Context) ([]Secret, error)
	}

	ListSecretParams struct {
		generics.ListParams
		FolderIDs  []uint
		Name       string
		Scriptable uint
	}

	// multiSorter implements the Sort interface, sorting the secrets within.
	multiSorter struct {
		secrets []*Secret
		less    []lessFunc
	}
)
