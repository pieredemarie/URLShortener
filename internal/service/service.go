package service

import (
	"URLShortener/internal/generator"
	"context"
	"net/url"
)

type URLRepository interface {
	Create(ctx context.Context, longURL string) (uint64, error)

	UpdateShortCode(ctx context.Context, id uint64, shortCode string) error

	GetLongLink(ctx context.Context, shortCode string) (string, error)

	GetShortLink(ctx context.Context, longURL string) (string, error)
}

type CacheRepository interface {
	GetLongLink(ctx context.Context, shortCode string) (string, error)
	SetLongLink(ctx context.Context, shortCode, longURL string) error
}

type URLService interface {
	GetOrCreate(ctx context.Context, longUrl string) (string, error)
	GetLongLink(ctx context.Context, shortCode string) (string, error)
	IsValidUrl(longUrl string) bool
}

type service struct {
	repo  URLRepository
	cache CacheRepository
}

func (s *service) GetLongLink(ctx context.Context, shortCode string) (string, error) {
	longURL, err := s.cache.GetLongLink(ctx, shortCode)
	if err == nil {
		return longURL, nil
	}

	longURL, err = s.repo.GetLongLink(ctx, shortCode)
	if err != nil {
		return "", err
	}

	_ = s.cache.SetLongLink(ctx, shortCode, longURL)

	return longURL, nil
}

func (s *service) GetOrCreate(
	ctx context.Context,
	longURL string,
) (string, error) {

	shortCode, err := s.repo.GetShortLink(ctx, longURL)

	if err == nil {
		return shortCode, nil
	}

	id, err := s.repo.Create(ctx, longURL)

	if err != nil {
		return "", err
	}

	shortCode = generator.Encode(id)

	err = s.repo.UpdateShortCode(ctx, id, shortCode)

	if err != nil {
		return "", err
	}

	_ = s.cache.SetLongLink(ctx, longURL, shortCode)

	return shortCode, nil
}

func (s *service) IsValidUrl(longUrl string) bool {
	parsedUrl, err := url.Parse(longUrl)
	if err != nil {
		return false
	}

	if parsedUrl.Scheme != "http" && parsedUrl.Scheme != "https" {
		return false
	}

	if parsedUrl.Host == "" {
		return false
	}

	return true
}
