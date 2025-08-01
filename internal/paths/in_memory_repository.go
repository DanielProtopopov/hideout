package paths

import (
	"context"
	"database/sql"
	"hideout/internal/common/apperror"
	"hideout/internal/common/generics"
	"hideout/internal/common/model"
	"hideout/internal/common/ordering"
	"hideout/internal/common/pagination"
	pathPkg "path"
	"slices"
	"time"
)

type InMemoryRepository struct {
	conn *[]Path
}

func NewInMemoryRepository(conn *[]Path) *InMemoryRepository {
	return &InMemoryRepository{conn: conn}
}

func (m InMemoryRepository) Load(ctx context.Context) ([]Path, error) {
	return nil, nil
}

func (m InMemoryRepository) GetID(ctx context.Context) (uint, error) {
	id := uint(0)
	for _, pathEntry := range *m.conn {
		if pathEntry.ID > id {
			id = pathEntry.ID
		}
	}

	return id + 1, nil
}

func (m InMemoryRepository) GetMapByID(ctx context.Context, params ListPathParams) (map[uint]*Path, error) {
	paths, errGetPaths := m.Get(ctx, params)
	if errGetPaths != nil {
		return nil, errGetPaths
	}
	results := make(map[uint]*Path)
	for _, pathEntry := range paths {
		results[pathEntry.ID] = pathEntry
	}

	return results, nil
}

func (m InMemoryRepository) GetMapByUID(ctx context.Context, params ListPathParams) (map[string]*Path, error) {
	paths, errGetPaths := m.Get(ctx, params)
	if errGetPaths != nil {
		return nil, errGetPaths
	}
	results := make(map[string]*Path)
	for _, pathEntry := range paths {
		results[pathEntry.UID] = pathEntry
	}

	return results, nil
}

func (m InMemoryRepository) Get(ctx context.Context, params ListPathParams) ([]*Path, error) {
	var parentPathResults []*Path
	for _, pathEntry := range *m.conn {
		if params.ParentPathID > 0 {
			if pathEntry.ParentID == params.ParentPathID {
				parentPathResults = append(parentPathResults, &pathEntry)
			}
		} else {
			parentPathResults = append(parentPathResults, &pathEntry)
		}
	}

	var nameResults []*Path
	for _, pathEntry := range parentPathResults {
		if params.Name != "" {
			matched, errPathMatch := pathPkg.Match(params.Name, pathEntry.Name)
			if errPathMatch != nil {
				return nil, errPathMatch
			}
			if matched {
				nameResults = append(nameResults, pathEntry)
			}
		} else {
			nameResults = append(nameResults, pathEntry)
		}
	}

	filteredResults := m.Filter(ctx, nameResults, params.ListParams)
	if params.Page == 0 && params.PerPage == 0 {
		return filteredResults, nil
	}

	offset, length := pagination.Paginate(len(filteredResults), int(params.Page*params.PerPage), int(params.PerPage))
	return filteredResults[offset:length], nil
}

func (m InMemoryRepository) GetByID(ctx context.Context, id uint) (*Path, error) {
	for _, pathEntry := range *m.conn {
		if pathEntry.ID == id && !pathEntry.DeletedAt.Valid {
			return &pathEntry, nil
		}
	}

	return nil, apperror.ErrRecordNotFound
}

func (m InMemoryRepository) GetByUID(ctx context.Context, uid string) (*Path, error) {
	for _, pathEntry := range *m.conn {
		if pathEntry.UID == uid && !pathEntry.DeletedAt.Valid {
			return &pathEntry, nil
		}
	}

	return nil, apperror.ErrRecordNotFound
}

func (m InMemoryRepository) Update(ctx context.Context, path Path) (*Path, error) {
	for _, pathEntry := range *m.conn {
		if pathEntry.ID == path.ID && !pathEntry.DeletedAt.Valid {
			pathEntry.ParentID = path.ParentID
			pathEntry.Name = path.Name
			pathEntry.UpdatedAt = time.Now()
			return &pathEntry, nil
		}
	}

	return nil, apperror.ErrRecordNotFound
}

func (m InMemoryRepository) Create(ctx context.Context, path Path) (*Path, error) {
	for _, pathEntry := range *m.conn {
		if !pathEntry.DeletedAt.Valid && pathEntry.Name == path.Name && pathEntry.ParentID == path.ParentID {
			return nil, apperror.ErrAlreadyExists
		}
	}

	path.CreatedAt = time.Now()
	*m.conn = append(*m.conn, path)
	return &path, nil
}

func (m InMemoryRepository) Count(ctx context.Context, params ListPathParams) (uint, error) {
	// These are not needed when performing filtering and counting
	params.Pagination = pagination.Pagination{PerPage: 0, Page: 0}
	params.Order = []ordering.Order{}

	pathsList, errGetPaths := m.Get(ctx, params)
	if errGetPaths != nil {
		return 0, errGetPaths
	}

	return uint(len(pathsList)), nil
}

func (m InMemoryRepository) Delete(ctx context.Context, id uint, forceDelete bool) error {
	for pathIndex, pathEntry := range *m.conn {
		if pathEntry.ID == id {
			if forceDelete {
				*m.conn = slices.Delete(*m.conn, pathIndex, pathIndex+1)
			} else {
				pathEntry.DeletedAt = sql.NullTime{Valid: true, Time: time.Now()}
			}
			return nil
		}
	}

	return apperror.ErrRecordNotFound
}

func (m InMemoryRepository) Filter(ctx context.Context, results []*Path, params generics.ListParams) []*Path {
	var idResults []*Path
	for _, pathEntry := range results {
		if len(params.IDs) > 0 {
			if slices.Contains(params.IDs, pathEntry.ID) {
				idResults = append(idResults, pathEntry)
			}
		} else {
			idResults = append(idResults, pathEntry)
		}
	}

	var uidResults []*Path
	for _, pathEntry := range idResults {
		if len(params.UIDs) > 0 {
			if slices.Contains(params.UIDs, pathEntry.UID) {
				uidResults = append(uidResults, pathEntry)
			}
		} else {
			uidResults = append(uidResults, pathEntry)
		}
	}

	var softDeletedResults []*Path
	for _, pathEntry := range uidResults {
		if params.Deleted == model.Yes {
			if pathEntry.DeletedAt.Valid {
				softDeletedResults = append(softDeletedResults, pathEntry)
			}
		} else if params.Deleted == model.No {
			if !pathEntry.DeletedAt.Valid {
				softDeletedResults = append(softDeletedResults, pathEntry)
			}
		} else {
			softDeletedResults = append(softDeletedResults, pathEntry)
		}
	}

	return m.Sort(ctx, softDeletedResults, params.Order)
}

func (m InMemoryRepository) Sort(ctx context.Context, data []*Path, ordering []ordering.Order) []*Path {
	var orderParams []lessFunc
	for _, order := range ordering {
		columnMap, _ := OrderMap[order.OrderBy]
		switch columnMap {
		case "id":
			{
				if order.Order == true {
					orderParams = append(orderParams, func(p1, p2 *Path) bool {
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
					orderParams = append(orderParams, func(p1, p2 *Path) bool {
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
					orderParams = append(orderParams, func(p1, p2 *Path) bool {
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
					orderParams = append(orderParams, func(p1, p2 *Path) bool {
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
					orderParams = append(orderParams, func(p1, p2 *Path) bool {
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
					orderParams = append(orderParams, func(p1, p2 *Path) bool {
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
					orderParams = append(orderParams, func(p1, p2 *Path) bool {
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
