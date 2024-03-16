package server

import (
	"fmt"

	"github.com/abdheshnayak/kubewiremesh/controllers/constants"
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

	app.Post("/configs", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World!")
	})

	app.Get("/healthy", func(c *fiber.Ctx) error {
		return c.SendStatus(200)
	})

	app.Get("/*", func(c *fiber.Ctx) error {
		return c.SendStatus(404)
	})

	return app.Listen(fmt.Sprintf(":%d", constants.RECEIVE_PORT))
}

func New(mgr manager.Manager, logger logging.Logger) *Server {
	return &Server{manager: mgr, logger: logger}
}
