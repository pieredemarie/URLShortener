package service

import (
	"URLShortener/internal/generator"
	"context"
	"errors"
	"testing"
)

type mockURLRepository struct {
	shortCode string
	longURL   string
	id        uint64

	getShortLinkErr error
	createErr       error
	updateErr       error
	getLongLinkErr  error

	getShortLinkCalls int
	createCalls       int
	updateCalls       int
	getLongLinkCalls  int

	createdLongURL string
	updatedID      uint64
	updatedCode    string
}

func (m *mockURLRepository) Create(
	_ context.Context,
	longURL string,
) (uint64, error) {
	m.createCalls++
	m.createdLongURL = longURL

	if m.createErr != nil {
		return 0, m.createErr
	}

	return m.id, nil
}

func (m *mockURLRepository) UpdateShortCode(
	_ context.Context,
	id uint64,
	shortCode string,
) error {
	m.updateCalls++
	m.updatedID = id
	m.updatedCode = shortCode

	return m.updateErr
}

func (m *mockURLRepository) GetLongLink(
	_ context.Context,
	shortCode string,
) (string, error) {
	m.getLongLinkCalls++

	if m.getLongLinkErr != nil {
		return "", m.getLongLinkErr
	}

	if shortCode != m.shortCode {
		return "", errors.New("not found")
	}

	return m.longURL, nil
}

func (m *mockURLRepository) GetShortLink(
	_ context.Context,
	longURL string,
) (string, error) {
	m.getShortLinkCalls++

	if m.getShortLinkErr != nil {
		return "", m.getShortLinkErr
	}

	if longURL != m.longURL {
		return "", errors.New("not found")
	}

	return m.shortCode, nil
}

type mockCacheRepository struct {
	longURL string
	err     error

	getCalls int
	setCalls int

	setShortCode string
	setLongURL   string
	setErr       error
}

func (m *mockCacheRepository) GetLongLink(
	_ context.Context,
	shortCode string,
) (string, error) {
	m.getCalls++

	if m.err != nil {
		return "", m.err
	}

	if shortCode != m.setShortCode {
		return "", errors.New("not found in cache")
	}

	return m.longURL, nil
}

func (m *mockCacheRepository) SetLongLink(
	_ context.Context,
	shortCode, longURL string,
) error {
	m.setCalls++
	m.setShortCode = shortCode
	m.setLongURL = longURL

	return m.setErr
}

func TestGetOrCreate_ExistingURL(t *testing.T) {
	repo := &mockURLRepository{
		shortCode:       "abc",
		longURL:         "https://example.com",
		getShortLinkErr: nil,
	}
	cache := &mockCacheRepository{}

	svc := New(repo, cache)

	got, err := svc.GetOrCreate(
		context.Background(),
		"https://example.com",
	)
	if err != nil {
		t.Fatalf("GetOrCreate() error = %v", err)
	}

	if got != "abc" {
		t.Fatalf("GetOrCreate() = %q, want %q", got, "abc")
	}

	if repo.createCalls != 0 {
		t.Errorf("Create() called %d times, want 0", repo.createCalls)
	}

	if repo.updateCalls != 0 {
		t.Errorf("UpdateShortCode() called %d times, want 0", repo.updateCalls)
	}
}

func TestGetOrCreate_NewURL(t *testing.T) {
	repo := &mockURLRepository{
		getShortLinkErr: errors.New("not found"),
		id:              42,
	}
	cache := &mockCacheRepository{}

	svc := New(repo, cache)

	longURL := "https://example.com"
	got, err := svc.GetOrCreate(context.Background(), longURL)
	if err != nil {
		t.Fatalf("GetOrCreate() error = %v", err)
	}

	want := generator.Encode(42)
	if got != want {
		t.Fatalf("GetOrCreate() = %q, want %q", got, want)
	}

	if repo.createCalls != 1 {
		t.Errorf("Create() called %d times, want 1", repo.createCalls)
	}

	if repo.createdLongURL != longURL {
		t.Errorf("Create() URL = %q, want %q", repo.createdLongURL, longURL)
	}

	if repo.updateCalls != 1 {
		t.Fatalf("UpdateShortCode() called %d times, want 1", repo.updateCalls)
	}

	if repo.updatedID != 42 {
		t.Errorf("updated ID = %d, want 42", repo.updatedID)
	}

	if repo.updatedCode != want {
		t.Errorf("updated code = %q, want %q", repo.updatedCode, want)
	}

	// Проверяем контракт кеша: ключ — короткий код,
	// значение — исходный URL.
	if cache.setShortCode != want {
		t.Errorf(
			"cache key = %q, want short code %q",
			cache.setShortCode,
			want,
		)
	}

	if cache.setLongURL != longURL {
		t.Errorf(
			"cache value = %q, want long URL %q",
			cache.setLongURL,
			longURL,
		)
	}
}

func TestGetOrCreate_CreateError(t *testing.T) {
	wantErr := errors.New("database unavailable")

	repo := &mockURLRepository{
		getShortLinkErr: errors.New("not found"),
		createErr:       wantErr,
	}
	cache := &mockCacheRepository{}

	svc := New(repo, cache)

	_, err := svc.GetOrCreate(
		context.Background(),
		"https://example.com",
	)
	if !errors.Is(err, wantErr) {
		t.Fatalf("GetOrCreate() error = %v, want %v", err, wantErr)
	}

	if repo.updateCalls != 0 {
		t.Errorf("UpdateShortCode() called %d times, want 0", repo.updateCalls)
	}

	if cache.setCalls != 0 {
		t.Errorf("SetLongLink() called %d times, want 0", cache.setCalls)
	}
}

func TestGetOrCreate_UpdateError(t *testing.T) {
	wantErr := errors.New("update failed")

	repo := &mockURLRepository{
		getShortLinkErr: errors.New("not found"),
		id:              42,
		updateErr:       wantErr,
	}
	cache := &mockCacheRepository{}

	svc := New(repo, cache)

	_, err := svc.GetOrCreate(
		context.Background(),
		"https://example.com",
	)
	if !errors.Is(err, wantErr) {
		t.Fatalf("GetOrCreate() error = %v, want %v", err, wantErr)
	}

	if cache.setCalls != 0 {
		t.Errorf("SetLongLink() called %d times, want 0", cache.setCalls)
	}
}

func TestGetLongLink_CacheHit(t *testing.T) {
	repo := &mockURLRepository{}
	cache := &mockCacheRepository{
		setShortCode: "abc",
		longURL:      "https://example.com",
	}

	svc := New(repo, cache)

	got, err := svc.GetLongLink(context.Background(), "abc")
	if err != nil {
		t.Fatalf("GetLongLink() error = %v", err)
	}

	if got != "https://example.com" {
		t.Fatalf("GetLongLink() = %q, want %q", got, "https://example.com")
	}

	if repo.getLongLinkCalls != 0 {
		t.Errorf("repository called %d times, want 0", repo.getLongLinkCalls)
	}
}

func TestGetLongLink_CacheMiss(t *testing.T) {
	repo := &mockURLRepository{
		shortCode: "abc",
		longURL:   "https://example.com",
	}
	cache := &mockCacheRepository{
		err: errors.New("cache miss"),
	}

	svc := New(repo, cache)

	got, err := svc.GetLongLink(context.Background(), "abc")
	if err != nil {
		t.Fatalf("GetLongLink() error = %v", err)
	}

	if got != "https://example.com" {
		t.Fatalf("GetLongLink() = %q, want %q", got, "https://example.com")
	}

	if repo.getLongLinkCalls != 1 {
		t.Errorf("repository called %d times, want 1", repo.getLongLinkCalls)
	}

	if cache.setCalls != 1 {
		t.Errorf("SetLongLink() called %d times, want 1", cache.setCalls)
	}

	if cache.setShortCode != "abc" {
		t.Errorf("cache key = %q, want %q", cache.setShortCode, "abc")
	}

	if cache.setLongURL != "https://example.com" {
		t.Errorf(
			"cache value = %q, want %q",
			cache.setLongURL,
			"https://example.com",
		)
	}
}

func TestGetLongLink_RepositoryError(t *testing.T) {
	wantErr := errors.New("database unavailable")

	repo := &mockURLRepository{
		getLongLinkErr: wantErr,
	}
	cache := &mockCacheRepository{
		err: errors.New("cache miss"),
	}

	svc := New(repo, cache)

	_, err := svc.GetLongLink(context.Background(), "abc")
	if !errors.Is(err, wantErr) {
		t.Fatalf("GetLongLink() error = %v, want %v", err, wantErr)
	}

	if cache.setCalls != 0 {
		t.Errorf("SetLongLink() called %d times, want 0", cache.setCalls)
	}
}

func TestIsValidUrl(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want bool
	}{
		{
			name: "valid HTTP",
			url:  "http://example.com",
			want: true,
		},
		{
			name: "valid HTTPS",
			url:  "https://example.com/path",
			want: true,
		},
		{
			name: "empty URL",
			url:  "",
			want: false,
		},
		{
			name: "missing scheme",
			url:  "example.com",
			want: false,
		},
		{
			name: "unsupported scheme",
			url:  "ftp://example.com",
			want: false,
		},
		{
			name: "missing host",
			url:  "https:///path",
			want: false,
		},
	}

	svc := New(&mockURLRepository{}, &mockCacheRepository{})

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := svc.IsValidUrl(tt.url)
			if got != tt.want {
				t.Errorf("IsValidUrl(%q) = %v, want %v", tt.url, got, tt.want)
			}
		})
	}
}
