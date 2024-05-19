package apiserver

import (
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

type Server struct {
	FiberApp *fiber.App
	DB       *gorm.DB
}

func New(db *gorm.DB, fiberApp *fiber.App) *Server {
	return &Server{FiberApp: fiberApp, DB: db}
}

func (s *Server) SetupRoutes() {
	s.FiberApp.Post("/accounts", s.CreateAccount)
	s.FiberApp.Get("/accounts/:account_id", s.GetAccount)
	s.FiberApp.Post("/transactions", s.CreateTransfer)
}

func (s *Server) Start(address string) error {
	return s.FiberApp.Listen(address)
}
