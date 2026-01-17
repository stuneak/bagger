package api

import (
	"github.com/gin-gonic/gin"
	db "github.com/stuneak/bagger/db/sqlc"
)

type Server struct {
	store  *db.Queries
	router *gin.Engine
}

func NewServer(store *db.Queries) *Server {
	server := &Server{store: store}
	router := gin.Default()

	// Routes
	router.GET("/users/:id", server.getUser)
	router.POST("/users", server.createUser)
	router.GET("/users", server.listUsers)

	server.router = router
	return server
}

func (server *Server) Start(address string) error {
	return server.router.Run(address)
}
