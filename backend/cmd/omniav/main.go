package main

import (
	"log/slog"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/ofgrenudo/OmniAv/internal/handlers"
	"github.com/ofgrenudo/OmniAv/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	app := models.App{}
	if err := app.DB.Parse(); err != nil {
		slog.Error("Failed to parse db config", slog.Any("error", err))
		os.Exit(1)
	}

	db, err := gorm.Open(postgres.Open(app.DB.DSN()), &gorm.Config{})
	if err != nil {
		slog.Error("Failed to connect to database", slog.Any("error", err))
		os.Exit(1)
	}

	if err := db.AutoMigrate(
		&models.Building{},
		&models.BuildingRoom{},
		&models.EquipmentGroup{},
		&models.Equipment{},
		&models.Request{},
		&models.RequestedEquipment{},
	); err != nil {
		slog.Error("Failed to run migrations", slog.Any("error", err))
		os.Exit(1)
	}

	slog.Info("Database migrations completed successfully")

	r := gin.Default()

	api := r.Group("/api")
	handlers.NewBuildingHandler(db).RegisterRoutes(api)
	handlers.NewEquipmentGroupHandler(db).RegisterRoutes(api)
	handlers.NewEquipmentHandler(db).RegisterRoutes(api)
	handlers.NewRequestHandler(db).RegisterRoutes(api)
	handlers.NewRequestedEquipmentHandler(db).RegisterRoutes(api)

	r.Run("0.0.0.0:8001")
}
