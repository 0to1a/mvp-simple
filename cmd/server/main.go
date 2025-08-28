package main

import (
	"log"

	core "project/internal"
	"project/internal/api"
	"project/internal/db"
)

func main() {
	cfg := core.LoadConfig()

	sqlDB, err := db.OpenAndMigrate(cfg.DatabaseDSN)
	if err != nil {
		log.Fatalf("db init/migrate: %v", err)
	}
	defer sqlDB.Close()

	app := core.NewApp(cfg, sqlDB)
	r := api.Build(app)

	port := "8080" // Hardcoded port
	log.Printf("listening on :%s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}
