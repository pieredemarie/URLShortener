package handlers

import (
	"URLShortener/internal/dto"
	"URLShortener/internal/service"
	"encoding/json"
	"net/http"
	"strings"
)

type Handler struct {
	service service.URLService
}

func New(service service.URLService) *Handler {
	return &Handler{
		service: service,
	}
}

func (h *Handler) Shorten(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req dto.UrlRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "bad payload", http.StatusBadRequest) // it'll be necessary to make custom errors for this?
		return
	}

	if req.URL == "" {
		http.Error(w, "url is required", http.StatusBadRequest)
		return
	}

	//some function to validate url

	shortUrl, err := h.service.GetOrCreate(ctx, req.URL)
	if err != nil {
		http.Error(w, "server error", http.StatusInternalServerError)
		return
	}

	var resp dto.UrlResponse
	resp.URL = shortUrl

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"short-url": resp.URL})
}

func (h *Handler) Redirect(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	path := strings.TrimPrefix(r.URL.Path, "/")

	shortCode := strings.Trim(path, "/")

	if shortCode == "" {
		http.Error(w, "short url is required", http.StatusBadRequest)
		return
	}

	longURL, err := h.service.GetLongLink(ctx, shortCode)
	if err != nil {
		http.Error(w, "URL not found", http.StatusNotFound)
		return
	}

	http.Redirect(w, r, longURL, http.StatusMovedPermanently)
}
