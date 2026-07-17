package service

import "context"

type URLService interface {
	CreateShortLink(ctx context.Context, shortLink, longURL string) error
	GetLongLink(ctx context.Context, shortUrl string) (string, error)

	// GetOrCreate if link exists return its shortURL, if not - create and return
	GetOrCreate(ctx context.Context, longURL string) (string, error)
}
