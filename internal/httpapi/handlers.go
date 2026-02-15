package httpapi

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"urlshortener/internal/repository/postgres"
	"urlshortener/internal/service"
)

type API struct {
	shortener *service.Shortener
}

type shortenRequest struct {
	URL string `json:"url"`
}

type shortenResponse struct {
	Code     string `json:"code"`
	ShortURL string `json:"short_url"`
}

func NewAPI(shortener *service.Shortener) *API {
	return &API{shortener: shortener}
}

func (a *API) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("GET /healthz", a.handleHealthz)
	mux.HandleFunc("POST /shorten", a.handleShorten)
	mux.HandleFunc("GET /{code}", a.handleRedirect)
}

func (a *API) handleHealthz(w http.ResponseWriter, _ *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (a *API) handleShorten(w http.ResponseWriter, r *http.Request) {
	var req shortenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	code, shortURL, err := a.shortener.CreateShortURL(r.Context(), req.URL)
	if errors.Is(err, service.ErrInvalidURL) {
		http.Error(w, "invalid URL", http.StatusBadRequest)
		return
	}
	if err != nil {
		log.Printf("shorten failed: %v", err)
		http.Error(w, "failed to shorten URL", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusCreated, shortenResponse{Code: code, ShortURL: shortURL})
}

func (a *API) handleRedirect(w http.ResponseWriter, r *http.Request) {
	code := r.PathValue("code")
	originalURL, err := a.shortener.ResolveCode(r.Context(), code)
	if errors.Is(err, postgres.ErrNotFound) {
		http.NotFound(w, r)
		return
	}
	if err != nil {
		log.Printf("resolve failed for code=%s: %v", code, err)
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, originalURL, http.StatusFound)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
