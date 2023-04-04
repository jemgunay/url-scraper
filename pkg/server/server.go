package server

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	"jemgunay/url-scraper/pkg/config"
	"jemgunay/url-scraper/pkg/ports"
)

type Server struct {
	logger  config.Logger
	addr    string
	scraper ports.Scraper
	storage ports.Storer

	router *gin.Engine
}

func New(logger config.Logger, port int, scraper ports.Scraper, storage ports.Storer) *Server {
	router := gin.Default()
	server := &Server{
		logger:  logger,
		addr:    fmt.Sprintf(":%d", port),
		scraper: scraper,
		storage: storage,

		router: router,
	}

	versionRouter := router.Any("/v1")
	apiRouter := versionRouter.Any("/api")
	apiRouter.GET("/urls", server.GetURL)
	apiRouter.POST("/urls", server.AddURL)

	return server
}

func (s *Server) Run() error {
	return s.router.Run(s.addr)
}

func (s *Server) GetURL(c *gin.Context) {
	ctx := c.Request.Context()

	// TODO: read directly from storage
	s.storage.Fetch()

	// preview, err := a.blog.GetPostPreviews(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch posts"})
		return
	}

	c.JSON(http.StatusOK, "")
}

func (s *Server) AddURL(c *gin.Context) {
	ctx := c.Request.Context()

	// TODO: add vie ingester
	s.scraper.
}
