package paths

import (
	"context"
	"database/sql"
	"github.com/brianvoe/gofakeit/v7"
	"hideout/internal/common/generics"
	"hideout/internal/common/model"
	error2 "hideout/internal/pkg/error"
	pathPkg "path"
	"slices"
	"strings"
	"time"
)

type InMemoryRepository struct {
	conn *[]Path
}

func NewInMemoryRepository(conn *[]Path) InMemoryRepository {
	return InMemoryRepository{conn: conn}
}

func (m InMemoryRepository) getID() uint {
	id := uint(0)
	for _, pathEntry := range *m.conn {
		if pathEntry.ID > id {
			id = pathEntry.ID
		}
	}

	return id + 1
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

	return m.Filter(ctx, nameResults, params.ListParams)
}

func (m InMemoryRepository) GetByID(ctx context.Context, id uint) (*Path, error) {
	for _, pathEntry := range *m.conn {
		if pathEntry.ID == id && !pathEntry.DeletedAt.Valid {
			return &pathEntry, nil
		}
	}

	return nil, error2.ErrRecordNotFound
}

func (m InMemoryRepository) GetByUID(ctx context.Context, uid string) (*Path, error) {
	for _, pathEntry := range *m.conn {
		if pathEntry.UID == uid && !pathEntry.DeletedAt.Valid {
			return &pathEntry, nil
		}
	}

	return nil, error2.ErrRecordNotFound
}

func (m InMemoryRepository) Update(ctx context.Context, id uint, name string) (*Path, error) {
	for _, pathEntry := range *m.conn {
		if pathEntry.ID == id && !pathEntry.DeletedAt.Valid {
			pathEntry.Name = name
			return &pathEntry, nil
		}
	}

	return nil, error2.ErrRecordNotFound
}

func (m InMemoryRepository) Create(ctx context.Context, parentPathID uint, name string) (*Path, error) {
	for _, pathEntry := range *m.conn {
		if pathEntry.Name == name && pathEntry.ID == parentPathID && !pathEntry.DeletedAt.Valid {
			return nil, error2.ErrAlreadyExists
		}
	}

	newPath := Path{Model: model.Model{ID: m.getID()}, UID: gofakeit.UUID(), ParentID: parentPathID, Name: name}
	*m.conn = append(*m.conn, newPath)
	return &newPath, nil
}

func (m InMemoryRepository) Count(ctx context.Context, name string) (uint, error) {
	totalCount := uint(0)
	for _, pathEntry := range *m.conn {
		if strings.Contains(pathEntry.Name, name) && !pathEntry.DeletedAt.Valid {
			totalCount++
		}
	}

	return totalCount, nil
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

	return error2.ErrRecordNotFound
}

func (m InMemoryRepository) Filter(ctx context.Context, results []*Path, params generics.ListParams) ([]*Path, error) {
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

	return uidResults, nil
}
