package dto

type UrlRequest struct {
	URL string `json:"url"`
}

type UrlResponse struct {
	URL string `json:"short_url"`
}
