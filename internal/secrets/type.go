package secrets

import "hideout/internal/common/generics"

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

	ListSecretParams struct {
		generics.ListParams
		Path  string
		Name  string
		Types []uint
	}
)
