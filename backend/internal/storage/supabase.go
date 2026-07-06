package storage

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// SupabaseStorage implements Provider for Supabase Storage
type SupabaseStorage struct {
	url           string // Base Supabase URL (e.g., https://project.supabase.co)
	serviceKey    string // Service role key (private, for backend use)
	bucketName    string // Bucket name in Supabase Storage
	httpClient    *http.Client
	maxUploadSize int64 // Maximum upload size in bytes
	pathValidator *ObjectPathValidator
}

// NewSupabaseStorage creates a new Supabase storage provider
// url: Supabase project URL (e.g., "https://project.supabase.co")
// serviceKey: Service role key (backend-only, never expose)
// bucketName: Bucket name in Supabase Storage (e.g., "media")
// maxUploadSize: Maximum upload size in bytes (e.g., 10*1024*1024 for 10MB). Use 0 for default 10MB.
func NewSupabaseStorage(url, serviceKey, bucketName string, maxUploadSize int64) (*SupabaseStorage, error) {
	if strings.TrimSpace(url) == "" {
		return nil, fmt.Errorf("supabase URL is required")
	}
	if strings.TrimSpace(serviceKey) == "" {
		return nil, fmt.Errorf("supabase service key is required")
	}
	if strings.TrimSpace(bucketName) == "" {
		return nil, fmt.Errorf("supabase bucket name is required")
	}

	// Default max upload size: 10MB
	if maxUploadSize <= 0 {
		maxUploadSize = 10 * 1024 * 1024
	}

	// Normalize URL - remove trailing slash
	url = strings.TrimSuffix(url, "/")

	return &SupabaseStorage{
		url:           url,
		serviceKey:    serviceKey,
		bucketName:    bucketName,
		httpClient:    &http.Client{},
		maxUploadSize: maxUploadSize,
		pathValidator: NewObjectPathValidator(512),
	}, nil
}

// Upload stores a file in Supabase Storage and returns the public URL
func (s *SupabaseStorage) Upload(ctx context.Context, objectPath string, content io.Reader, contentType string) (string, error) {
	// Validate objectPath
	if err := s.pathValidator.Validate(objectPath); err != nil {
		return "", err
	}

	// Read content with size limit to prevent memory exhaustion
	limitedReader := io.LimitReader(content, s.maxUploadSize+1)
	data, err := io.ReadAll(limitedReader)
	if err != nil {
		return "", fmt.Errorf("failed to read file content: %w", err)
	}

	// Check if content exceeded max size
	if int64(len(data)) > s.maxUploadSize {
		return "", fmt.Errorf("file exceeds maximum upload size of %d bytes", s.maxUploadSize)
	}

	// Safely encode objectPath for URL
	safeObjectPath := s.pathValidator.SafeURL(objectPath)

	// Build upload URL
	// Format: {url}/storage/v1/object/{bucketName}/{safeObjectPath}
	uploadURL := fmt.Sprintf("%s/storage/v1/object/%s/%s", s.url, s.bucketName, safeObjectPath)

	// Create request
	req, err := http.NewRequestWithContext(ctx, "POST", uploadURL, bytes.NewReader(data))
	if err != nil {
		return "", fmt.Errorf("failed to create upload request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.serviceKey))
	req.Header.Set("Content-Type", contentType)

	// Execute upload
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("upload request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse response (Supabase returns JSON with file info)
	var uploadResp struct {
		Name string `json:"name"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&uploadResp); err != nil {
		// If response doesn't contain JSON, still consider it success if status is 2xx
		// Just use the objectPath for URL construction
	}

	// Build public URL - use original (unencoded) objectPath for display URL
	// Note: Supabase URL decodes path automatically, so we use original here
	publicURL := fmt.Sprintf("%s/storage/v1/object/public/%s/%s", s.url, s.bucketName, safeObjectPath)

	return publicURL, nil
}

// Delete removes a file from Supabase Storage
func (s *SupabaseStorage) Delete(ctx context.Context, objectPath string) error {
	// Validate objectPath
	if err := s.pathValidator.Validate(objectPath); err != nil {
		return err
	}

	// Safely encode objectPath for URL
	safeObjectPath := s.pathValidator.SafeURL(objectPath)

	// Build delete URL
	// Format: {url}/storage/v1/object/{bucketName}/{safeObjectPath}
	deleteURL := fmt.Sprintf("%s/storage/v1/object/%s/%s", s.url, s.bucketName, safeObjectPath)

	// Create request
	req, err := http.NewRequestWithContext(ctx, "DELETE", deleteURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create delete request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.serviceKey))

	// Execute delete
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("delete request failed: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	// Note: Supabase returns 204 No Content on successful delete
	// Also accept 200 OK as valid success
	if resp.StatusCode == 404 {
		// File not found - treat as success (idempotent delete)
		return nil
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete failed with status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	return nil
}

// Download reads a file from Supabase Storage with a hard byte limit.
func (s *SupabaseStorage) Download(ctx context.Context, objectPath string, maxBytes int64) ([]byte, error) {
	if maxBytes <= 0 {
		return nil, ErrFileTooLarge
	}
	if err := s.pathValidator.Validate(objectPath); err != nil {
		return nil, err
	}

	safeObjectPath := s.pathValidator.SafeURL(objectPath)
	downloadURL := fmt.Sprintf("%s/storage/v1/object/%s/%s", s.url, s.bucketName, safeObjectPath)

	req, err := http.NewRequestWithContext(ctx, "GET", downloadURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create download request: %w", err)
	}
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.serviceKey))

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("download request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, ErrNotFound
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	limitedReader := io.LimitReader(resp.Body, maxBytes+1)
	data, err := io.ReadAll(limitedReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read downloaded file: %w", err)
	}
	if int64(len(data)) > maxBytes {
		return nil, ErrFileTooLarge
	}
	return data, nil
}

// HealthCheck verifies Supabase Storage is available
func (s *SupabaseStorage) HealthCheck(ctx context.Context) error {
	// Simple health check: try to list bucket (minimal operation)
	// Format: {url}/storage/v1/bucket/{bucketName}
	checkURL := fmt.Sprintf("%s/storage/v1/bucket/%s", s.url, s.bucketName)

	// Create request
	req, err := http.NewRequestWithContext(ctx, "GET", checkURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create health check request: %w", err)
	}

	// Set headers
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", s.serviceKey))

	// Execute check
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("health check failed with status %d", resp.StatusCode)
	}

	return nil
}

// GetPublicURL returns the public URL for a file in storage
// This can be used for files that are already stored
func (s *SupabaseStorage) GetPublicURL(objectPath string) string {
	if objectPath == "" {
		return ""
	}
	// Safely encode objectPath for URL
	safeObjectPath := s.pathValidator.SafeURL(objectPath)
	return fmt.Sprintf("%s/storage/v1/object/public/%s/%s", s.url, s.bucketName, safeObjectPath)
}
