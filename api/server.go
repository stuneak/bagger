package api

import (
	"context"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	db "github.com/stuneak/sopeko/db/sqlc"
)

type Server struct {
	store  *db.Queries
	router *gin.Engine
}

func NewServer(store *db.Queries, ginMode string) *Server {
	server := &Server{store: store}
	router := gin.Default()

	gin.SetMode(ginMode)

	// CORS middleware - allow requests from browser extensions
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type"},
		AllowCredentials: false,
	}))

	// Visitor tracking middleware
	router.Use(server.visitorTrackingMiddleware())

	// Health check
	router.GET("/api/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Routes
	router.GET("/api/mentions/:username", server.getUserMentions)
	router.GET("/api/excluded-usernames", server.getExcludedUsernames)
	router.GET("/api/top-performers", server.getTopPerformingUsers)
	router.GET("/api/top-picks", server.getTopPerformingPicks)
	router.GET("/api/worst-picks", server.getWorstPerformingPicks)
	// router.GET("/api/visitors", server.getVisitorStats)
	server.router = router
	return server
}

func (server *Server) Start(address string) error {
	return server.router.Run(address)
}

func (server *Server) visitorTrackingMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		endpoint := c.Request.URL.Path

		go func() {
			_ = server.store.CreateVisitor(context.Background(), db.CreateVisitorParams{
				IpAddress: ip,
				Endpoint:  endpoint,
				VisitedAt: time.Now(),
			})
		}()

		c.Next()
	}
}
