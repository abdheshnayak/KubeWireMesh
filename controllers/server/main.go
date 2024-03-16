package server

import (
	"github.com/gofiber/fiber/v2"
	"github.com/kloudlite/operator/pkg/logging"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

type Server struct {
	manager manager.Manager
	logger  logging.Logger
}

func (s *Server) Run() error {
	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	return app.Listen(":3000")
}

func New(mgr manager.Manager, logger logging.Logger) *Server {
	return &Server{manager: mgr, logger: logger}
}
