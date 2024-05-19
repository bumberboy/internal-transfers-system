package main

import (
	"github.com/gofiber/fiber/v2"
	"internal-transfers-system/config"
	"internal-transfers-system/internal/apiserver"
	"internal-transfers-system/internal/database"
	"log"
)

func main() {
	conf, err := config.LoadConfig("app")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	db := database.NewDefaultDBClientOrFatal(conf)

	app := fiber.New()

	svr := apiserver.New(db, app)

	svr.SetupRoutes()
	log.Fatal(svr.Start(conf.SvrAddress))
}
