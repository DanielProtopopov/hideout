package folders

import (
	"context"
	"hideout/internal/common/apperror"

	"github.com/go-git/go-git/v6"
	"gorm.io/gorm"
)

type GitRepository struct {
	URL  string
	conn *git.Repository
}

func NewGitRepository(repository *git.Repository) (GitRepository, error) {
	return GitRepository{conn: repository}, nil
}

func (m GitRepository) GetID(ctx context.Context) (string, error) {
	return "", apperror.ErrNotImplemented
}

func (m GitRepository) Load(ctx context.Context) ([]Folder, error) {
	return []Folder{}, apperror.ErrNotImplemented
}

func (m GitRepository) GetMapByID(ctx context.Context, params ListFolderParams) (map[string]*Folder, error) {
	return map[string]*Folder{}, apperror.ErrNotImplemented
}

func (m GitRepository) GetMapByUID(ctx context.Context, params ListFolderParams) (map[string]*Folder, error) {
	return map[string]*Folder{}, apperror.ErrNotImplemented
}

func (m GitRepository) Get(ctx context.Context, params ListFolderParams) ([]*Folder, error) {
	return []*Folder{}, apperror.ErrNotImplemented
}

func (m GitRepository) GetMapByFolder(ctx context.Context, params ListFolderParams) (map[uint][]*Folder, error) {
	return map[uint][]*Folder{}, apperror.ErrNotImplemented
}

func (m GitRepository) GetByID(ctx context.Context, id uint) (*Folder, error) {
	return nil, apperror.ErrNotImplemented
}

func (m GitRepository) GetByUID(ctx context.Context, uid string) (*Folder, error) {
	return nil, apperror.ErrNotImplemented
}

func (m GitRepository) Update(ctx context.Context, folder Folder) (*Folder, error) {
	return nil, apperror.ErrNotImplemented
}

func (m GitRepository) Create(ctx context.Context, folder Folder) (*Folder, error) {
	return nil, apperror.ErrNotImplemented
}

func (m GitRepository) Count(ctx context.Context, params ListFolderParams) (uint, error) {
	return 0, apperror.ErrNotImplemented
}

func (m GitRepository) Delete(ctx context.Context, id uint, forceDelete bool) error {
	return apperror.ErrNotImplemented
}

func (m GitRepository) GetQuery(tx *gorm.DB, selectedColumnNames []string, params ListFolderParams) (Query *gorm.DB) {
	return nil
}
