package storage

import (
	"context"
	"io"
)

// DisabledStorage is a no-op storage implementation that always returns ErrNotImplemented
// Used when storage is not yet implemented or disabled
type DisabledStorage struct{}

// NewDisabledStorage creates a disabled storage provider
func NewDisabledStorage() *DisabledStorage {
	return &DisabledStorage{}
}

// Upload returns ErrNotImplemented
func (d *DisabledStorage) Upload(ctx context.Context, objectPath string, content io.Reader, contentType string) (string, error) {
	return "", ErrNotImplemented
}

// Delete returns ErrNotImplemented
func (d *DisabledStorage) Delete(ctx context.Context, objectPath string) error {
	return ErrNotImplemented
}

// HealthCheck returns ErrUnavailable (storage not available)
func (d *DisabledStorage) HealthCheck(ctx context.Context) error {
	return ErrUnavailable
}

// GetPublicURL returns empty string (not implemented)
func (d *DisabledStorage) GetPublicURL(objectPath string) string {
	return ""
}
