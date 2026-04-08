package access_control

import (
	"context"
	"hideout/internal/common/apperror"
)

type AccessControlService struct{}

// NewService Creation of the service
func NewService(ctx context.Context) (*AccessControlService, error) {
	accessControlService := &AccessControlService{}
	return accessControlService, nil
}

// CountFolder check that the user has access to count entries in the folder
func (m *AccessControlService) CountFolder(ctx context.Context, folderUID string) error {
	return apperror.ErrNotImplemented
}

// ListFolder check that the user has access to list entries in the folder (only see key names)
func (m *AccessControlService) ListFolder(ctx context.Context, folderUID string) error {
	return apperror.ErrNotImplemented
}

// ReadSecret check that user has access to read secret name and value
func (m *AccessControlService) ReadSecret(ctx context.Context, folderUID string, secretUID string) error {
	return apperror.ErrNotImplemented
}

// WriteSecret check that the user has access to write secret value
func (m *AccessControlService) WriteSecret(ctx context.Context, folderUID string, secretUID string) error {
	return apperror.ErrNotImplemented
}

// DeleteSecret check that the user has access to delete secret
func (m *AccessControlService) DeleteSecret(ctx context.Context, folderUID string, secretUID string) error {
	return apperror.ErrNotImplemented
}
