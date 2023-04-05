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

// Server provides a RESTful HTTP server for performing URL-related operations.
type Server struct {
	logger   config.Logger
	ingester ports.Ingester
	storage  ports.Storer

	httpServer *http.Server
}

// New initialises a new HTTP URL API server.
func New(logger config.Logger, port int, ingester ports.Ingester, storage ports.Storer) *Server {
	server := &Server{
		logger:   logger,
		ingester: ingester,
		storage:  storage,
	}

	// disable gin debug logs
	gin.SetMode(gin.ReleaseMode)

	// register routes in router
	router := gin.Default()
	api := router.Group("/api")
	v1 := api.Group("/v1")
	v1.GET("/urls", server.GetURL)
	v1.POST("/urls", server.AddURL)

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
// error on shutdown (on success or failure).
func (s *Server) Run() error {
	return s.httpServer.ListenAndServe()
}

// GetURL fetches the 50 most recently stored URLs with their submission count
// in JSON form. It accepts query parameters for sort criteria (sortBy=age/
// count, default age) and sort order (sortOrder=asc/desc, default desc).
func (s *Server) GetURL(c *gin.Context) {
	sortBy := ports.Age
	sortByRaw, ok := c.GetQuery("sortBy")
	if ok {
		sortBy = ports.SortBy(sortByRaw)
		if err := sortBy.Validate(); err != nil {
			const msg = "invalid sortBy query param provided"
			s.logger.Error(msg, zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": msg})
			return
		}
	}

	sortOrder := ports.Descending
	sortOrderRaw, ok := c.GetQuery("sortOrder")
	if ok {
		sortOrder = ports.SortOrder(sortOrderRaw)
		if err := sortOrder.Validate(); err != nil {
			const msg = "invalid sortOrder query param provided"
			s.logger.Error(msg, zap.Error(err))
			c.JSON(http.StatusBadRequest, gin.H{"error": msg})
			return
		}
	}

	records := s.storage.Fetch(50, sortBy, sortOrder)
	c.JSON(http.StatusOK, records)
}

type addPayload struct {
	URL string `json:"url"`
}

// AddURL accepts a URL to insert into the store. The storage operation is
// asynchronous and successful storage is not guaranteed despite an Accepted
// response status code.
func (s *Server) AddURL(c *gin.Context) {
	// set a context timeout to prevent ingester backpressure from starving the
	// server
	ctx := c.Request.Context()
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
