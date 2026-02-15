package httpapi

import (
	"errors"
	"log"
	"net/http"

	"urlshortener/internal/repository/postgres"
	"urlshortener/internal/service"

	"github.com/gin-gonic/gin"
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

func (a *API) RegisterRoutes(router *gin.Engine) {
	router.GET("/healthz", a.handleHealthz)
	router.POST("/shorten", a.handleShorten)
	router.GET("/:code", a.handleRedirect)
}

func (a *API) handleHealthz(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

func (a *API) handleShorten(c *gin.Context) {
	var req shortenRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid JSON body",
		})
		return
	}

	code, shortURL, err := a.shortener.CreateShortURL(c.Request.Context(), req.URL)

	if errors.Is(err, service.ErrInvalidURL) {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid URL",
		})
		return
	}

	if err != nil {
		log.Printf("shorten failed: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to shorten URL",
		})
		return
	}

	c.JSON(http.StatusCreated, shortenResponse{
		Code:     code,
		ShortURL: shortURL,
	})
}

func (a *API) handleRedirect(c *gin.Context) {
	code := c.Param("code")

	originalURL, err := a.shortener.ResolveCode(c.Request.Context(), code)

	if errors.Is(err, postgres.ErrNotFound) {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "not found",
		})
		return
	}

	if err != nil {
		log.Printf("resolve failed for code=%s: %v", code, err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
		return
	}

	c.Redirect(http.StatusFound, originalURL)
}
