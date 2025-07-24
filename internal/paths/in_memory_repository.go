package paths

import (
	"context"
	"github.com/brianvoe/gofakeit/v7"
	error2 "hideout/internal/pkg/error"
	pathPkg "path"
	"strings"
)

type InMemoryRepository struct {
	conn []Path
}

func NewRepository(conn []Path) Repository {
	return InMemoryRepository{conn: conn}
}

func (m InMemoryRepository) getID() uint {
	id := uint(0)
	for _, pathEntry := range m.conn {
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
	for _, path := range paths {
		results[path.ID] = path
	}

	return results, nil
}

func (m InMemoryRepository) GetMapByUID(ctx context.Context, params ListPathParams) (map[string]*Path, error) {
	paths, errGetPaths := m.Get(ctx, params)
	if errGetPaths != nil {
		return nil, errGetPaths
	}
	results := make(map[string]*Path)
	for _, path := range paths {
		results[path.UID] = path
	}

	return results, nil
}

func (m InMemoryRepository) Get(ctx context.Context, params ListPathParams) ([]*Path, error) {

	var nameResults []*Path
	for _, pathResult := range m.conn {
		matched, errPathMatch := pathPkg.Match(params.Name, pathResult.Name)
		if errPathMatch != nil {
			return nil, errPathMatch
		}
		if matched {
			nameResults = append(nameResults, &pathResult)
		}
	}

	return nameResults, nil
}

func (m InMemoryRepository) GetByID(ctx context.Context, id uint) (*Path, error) {
	for _, uidPath := range m.conn {
		if uidPath.ID == id {
			return &uidPath, nil
		}
	}

	return nil, error2.ErrRecordNotFound
}

func (m InMemoryRepository) GetByUID(ctx context.Context, uid string) (*Path, error) {
	for _, uidPath := range m.conn {
		if uidPath.UID == uid {
			return &uidPath, nil
		}
	}

	return nil, error2.ErrRecordNotFound
}

func (m InMemoryRepository) Update(ctx context.Context, id uint, name string) (*Path, error) {
	for _, pathEntry := range m.conn {
		if pathEntry.ID == id {
			pathEntry.Name = name
			return &pathEntry, nil
		}
	}

	return nil, error2.ErrRecordNotFound
}

func (m InMemoryRepository) Create(ctx context.Context, pathID uint, name string) (*Path, error) {
	for _, pathEntry := range m.conn {
		if pathEntry.Name == name && pathEntry.ID == pathID {
			return nil, error2.ErrAlreadyExists
		}
	}

	newPath := Path{ID: m.getID(), UID: gofakeit.UUID(), Name: name}
	m.conn = append(m.conn, newPath)
	return &newPath, nil
}

func (m InMemoryRepository) Count(ctx context.Context, name string) (uint, error) {
	totalCount := uint(0)
	for _, pathEntry := range m.conn {
		if strings.Contains(pathEntry.Name, name) {
			totalCount++
		}
	}

	return totalCount, nil
}

func (m InMemoryRepository) Delete(ctx context.Context, id uint) error {
	for pathIndex, pathEntry := range m.conn {
		if pathEntry.ID == id {
			m.conn = append(m.conn[:pathIndex], m.conn[pathIndex+1:]...)
			return nil
		}
	}

	return error2.ErrRecordNotFound
}
