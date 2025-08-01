package folders

import (
	"context"
	"database/sql"
	"hideout/internal/common/apperror"
	"hideout/internal/common/generics"
	"hideout/internal/common/model"
	"hideout/internal/common/ordering"
	"hideout/internal/common/pagination"
	"path"
	"slices"
	"time"
)

type InMemoryRepository struct {
	conn *[]Folder
}

func NewInMemoryRepository(conn *[]Folder) *InMemoryRepository {
	return &InMemoryRepository{conn: conn}
}

func (m InMemoryRepository) Load(ctx context.Context) ([]Folder, error) {
	return nil, nil
}

func (m InMemoryRepository) GetID(ctx context.Context) (uint, error) {
	id := uint(0)
	for _, folderEntry := range *m.conn {
		if folderEntry.ID > id {
			id = folderEntry.ID
		}
	}

	return id + 1, nil
}

func (m InMemoryRepository) GetMapByID(ctx context.Context, params ListFolderParams) (map[uint]*Folder, error) {
	folders, errGetFolders := m.Get(ctx, params)
	if errGetFolders != nil {
		return nil, errGetFolders
	}
	results := make(map[uint]*Folder)
	for _, folderEntry := range folders {
		results[folderEntry.ID] = folderEntry
	}

	return results, nil
}

func (m InMemoryRepository) GetMapByUID(ctx context.Context, params ListFolderParams) (map[string]*Folder, error) {
	folders, errGetFolders := m.Get(ctx, params)
	if errGetFolders != nil {
		return nil, errGetFolders
	}
	results := make(map[string]*Folder)
	for _, folderEntry := range folders {
		results[folderEntry.UID] = folderEntry
	}

	return results, nil
}

func (m InMemoryRepository) Get(ctx context.Context, params ListFolderParams) ([]*Folder, error) {
	var parentFolderResults []*Folder
	for _, folderEntry := range *m.conn {
		if params.ParentFolderID > 0 {
			if folderEntry.ParentID == params.ParentFolderID {
				parentFolderResults = append(parentFolderResults, &folderEntry)
			}
		} else {
			parentFolderResults = append(parentFolderResults, &folderEntry)
		}
	}

	var nameResults []*Folder
	for _, folderEntry := range parentFolderResults {
		if params.Name != "" {
			matched, errFolderMatch := path.Match(params.Name, folderEntry.Name)
			if errFolderMatch != nil {
				return nil, errFolderMatch
			}
			if matched {
				nameResults = append(nameResults, folderEntry)
			}
		} else {
			nameResults = append(nameResults, folderEntry)
		}
	}

	filteredResults := m.Filter(ctx, nameResults, params.ListParams)
	if params.Page == 0 && params.PerPage == 0 {
		return filteredResults, nil
	}

	offset, length := pagination.Paginate(len(filteredResults), int(params.Page*params.PerPage), int(params.PerPage))
	return filteredResults[offset:length], nil
}

func (m InMemoryRepository) GetByID(ctx context.Context, id uint) (*Folder, error) {
	for _, folderEntry := range *m.conn {
		if folderEntry.ID == id && !folderEntry.DeletedAt.Valid {
			return &folderEntry, nil
		}
	}

	return nil, apperror.ErrRecordNotFound
}

func (m InMemoryRepository) GetByUID(ctx context.Context, uid string) (*Folder, error) {
	for _, folderEntry := range *m.conn {
		if folderEntry.UID == uid && !folderEntry.DeletedAt.Valid {
			return &folderEntry, nil
		}
	}

	return nil, apperror.ErrRecordNotFound
}

func (m InMemoryRepository) Update(ctx context.Context, folder Folder) (*Folder, error) {
	for _, folderEntry := range *m.conn {
		if folderEntry.ID == folder.ID && !folderEntry.DeletedAt.Valid {
			folderEntry.ParentID = folder.ParentID
			folderEntry.Name = folder.Name
			folderEntry.UpdatedAt = time.Now()
			return &folderEntry, nil
		}
	}

	return nil, apperror.ErrRecordNotFound
}

func (m InMemoryRepository) Create(ctx context.Context, folder Folder) (*Folder, error) {
	for _, folderEntry := range *m.conn {
		if !folderEntry.DeletedAt.Valid && folderEntry.Name == folder.Name && folderEntry.ParentID == folder.ParentID {
			return nil, apperror.ErrAlreadyExists
		}
	}

	folder.CreatedAt = time.Now()
	*m.conn = append(*m.conn, folder)
	return &folder, nil
}

func (m InMemoryRepository) Count(ctx context.Context, params ListFolderParams) (uint, error) {
	// These are not needed when performing filtering and counting
	params.Pagination = pagination.Pagination{PerPage: 0, Page: 0}
	params.Order = []ordering.Order{}

	foldersList, errGetFolders := m.Get(ctx, params)
	if errGetFolders != nil {
		return 0, errGetFolders
	}

	return uint(len(foldersList)), nil
}

func (m InMemoryRepository) Delete(ctx context.Context, id uint, forceDelete bool) error {
	for folderIndex, folderEntry := range *m.conn {
		if folderEntry.ID == id {
			if forceDelete {
				*m.conn = slices.Delete(*m.conn, folderIndex, folderIndex+1)
			} else {
				folderEntry.DeletedAt = sql.NullTime{Valid: true, Time: time.Now()}
			}
			return nil
		}
	}

	return apperror.ErrRecordNotFound
}

func (m InMemoryRepository) Filter(ctx context.Context, results []*Folder, params generics.ListParams) []*Folder {
	var idResults []*Folder
	for _, folderEntry := range results {
		if len(params.IDs) > 0 {
			if slices.Contains(params.IDs, folderEntry.ID) {
				idResults = append(idResults, folderEntry)
			}
		} else {
			idResults = append(idResults, folderEntry)
		}
	}

	var uidResults []*Folder
	for _, folderEntry := range idResults {
		if len(params.UIDs) > 0 {
			if slices.Contains(params.UIDs, folderEntry.UID) {
				uidResults = append(uidResults, folderEntry)
			}
		} else {
			uidResults = append(uidResults, folderEntry)
		}
	}

	var softDeletedResults []*Folder
	for _, folderEntry := range uidResults {
		if params.Deleted == model.Yes {
			if folderEntry.DeletedAt.Valid {
				softDeletedResults = append(softDeletedResults, folderEntry)
			}
		} else if params.Deleted == model.No {
			if !folderEntry.DeletedAt.Valid {
				softDeletedResults = append(softDeletedResults, folderEntry)
			}
		} else {
			softDeletedResults = append(softDeletedResults, folderEntry)
		}
	}

	return m.Sort(ctx, softDeletedResults, params.Order)
}

func (m InMemoryRepository) Sort(ctx context.Context, data []*Folder, ordering []ordering.Order) []*Folder {
	var orderParams []lessFunc
	for _, order := range ordering {
		columnMap, _ := OrderMap[order.OrderBy]
		switch columnMap {
		case "id":
			{
				if order.Order == true {
					orderParams = append(orderParams, func(p1, p2 *Folder) bool {
						if order.Order {
							return p1.ID < p2.ID
						} else {
							return p1.ID > p2.ID
						}
					})
				}
			}
		case "parent_id":
			{
				if order.Order == true {
					orderParams = append(orderParams, func(p1, p2 *Folder) bool {
						if order.Order {
							return p1.ParentID < p2.ParentID
						} else {
							return p1.ParentID > p2.ParentID
						}
					})
				}
			}
		case "uid":
			{
				if order.Order == true {
					orderParams = append(orderParams, func(p1, p2 *Folder) bool {
						if order.Order {
							return p1.UID < p2.UID
						} else {
							return p1.UID > p2.UID
						}
					})
				}
			}
		case "name":
			{
				if order.Order == true {
					orderParams = append(orderParams, func(p1, p2 *Folder) bool {
						if order.Order {
							return p1.Name < p2.Name
						} else {
							return p1.Name > p2.Name
						}
					})
				}
			}
		case "created_at":
			{
				if order.Order == true {
					orderParams = append(orderParams, func(p1, p2 *Folder) bool {
						if order.Order {
							return p1.CreatedAt.Before(p2.CreatedAt)
						} else {
							return p1.CreatedAt.After(p2.CreatedAt)
						}
					})
				}
			}
		case "updated_at":
			{
				if order.Order == true {
					orderParams = append(orderParams, func(p1, p2 *Folder) bool {
						if order.Order {
							return p1.UpdatedAt.Before(p2.UpdatedAt)
						} else {
							return p1.UpdatedAt.After(p2.UpdatedAt)
						}
					})
				}
			}
		case "deleted_at":
			{
				if order.Order == true {
					orderParams = append(orderParams, func(p1, p2 *Folder) bool {
						if order.Order {
							return p1.DeletedAt.Valid && p2.DeletedAt.Valid && p1.DeletedAt.Time.Before(p2.DeletedAt.Time)
						} else {
							return p1.DeletedAt.Valid && p2.DeletedAt.Valid && p1.DeletedAt.Time.After(p2.DeletedAt.Time)
						}
					})
				}
			}
		}
	}

	OrderedBy(orderParams...).Sort(data)
	return data
}
