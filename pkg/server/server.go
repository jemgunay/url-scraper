package server

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"jemgunay/url-scraper/pkg/config"
	"jemgunay/url-scraper/pkg/ports"
)

type Server struct {
	logger   config.Logger
	ingester ports.Ingester
	storage  ports.Storer

	httpServer *http.Server
}

func New(logger config.Logger, port int, ingester ports.Ingester, storage ports.Storer) *Server {
	server := &Server{
		logger:   logger,
		ingester: ingester,
		storage:  storage,
	}

	gin.SetMode(gin.ReleaseMode)

	// register routes in router
	router := gin.Default()
	v1 := router.Group("/v1")
	api := v1.Group("/api")
	api.GET("/urls", server.GetURL)
	api.POST("/urls", server.AddURL)

	server.httpServer = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: router,
		// set some sane HTTP server defaults
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}

	return server
}

// Run binds the HTTP server. It is a blocking operation and always returns an
// error on shutdown.
func (s *Server) Run() error {
	return s.httpServer.ListenAndServe()
}

func (s *Server) GetURL(c *gin.Context) {
	// TODO: read directly from storage, only 50
	records := s.storage.Fetch()

	// preview, err := a.blog.GetPostPreviews(ctx)
	// if err != nil {
	// 	s.logger.Error()
	// 	c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch posts"})
	// 	return
	// }

	c.JSON(http.StatusOK, records)
}

type addPayload struct {
	URL string `json:"url"`
}

func (s *Server) AddURL(c *gin.Context) {
	ctx := c.Request.Context()
	// set a context timeout in case ingestion is saturated/takes too long
	ctx, cancel := context.WithTimeout(ctx, time.Second*10)
	defer cancel()

	payload := addPayload{}
	if err := c.BindJSON(&payload); err != nil {
		s.logger.Error("failed to JSON decode add URL request body", zap.Error(err))
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request payload"})
		return
	}

	if err := s.ingester.Ingest(ctx, payload.URL); err != nil {
		s.logger.Error("failed to ingest URL", zap.Error(err))
		c.JSON(http.StatusInternalServerError, gin.H{"error": "unexpected error adding URL"})
		return
	}

	c.Status(http.StatusAccepted)
}
