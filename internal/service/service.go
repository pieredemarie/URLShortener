package service

import (
	"URLShortener/internal/generator"
	"context"
)

type URLRepository interface {
	Create(ctx context.Context, longURL string) (uint64, error)

	UpdateShortCode(ctx context.Context, id uint64, shortCode string) error

	GetLongLink(ctx context.Context, shortCode string) (string, error)

	GetShortLink(ctx context.Context, longURL string) (string, error)
}

type URLService interface {
	GetOrCreate(ctx context.Context, longUrl string) (string, error)
	GetLongLink(ctx context.Context, shortCode string) (string, error)
}

type service struct {
	repo URLRepository
}

func (s *service) GetLongLink(ctx context.Context, shortCode string) (string, error) {
	return s.repo.GetLongLink(ctx, shortCode)
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

	return shortCode, nil
}
