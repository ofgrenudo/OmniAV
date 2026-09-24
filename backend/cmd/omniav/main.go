package main

import (
	"fmt"
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

	if err := app.Auth.Parse(); err != nil {
		slog.Error("Failed to parse auth config", slog.Any("error", err))
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
		&models.User{},
	); err != nil {
		slog.Error("Failed to run migrations", slog.Any("error", err))
		os.Exit(1)
	}

	// Equipment.BuildingID gained a NOT NULL constraint; backfill any rows left over from before
	// that constraint existed so AutoMigrate's ALTER COLUMN doesn't fail against real data.
	if err := backfillEquipmentBuildingID(db); err != nil {
		slog.Error("Failed to backfill equipment building_id", slog.Any("error", err))
		os.Exit(1)
	}

	if err := db.AutoMigrate(
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
	authHandler, err := handlers.NewAuthHandler(db, app.Auth)
	if err != nil {
		slog.Error("Failed to configure auth", slog.Any("error", err))
		os.Exit(1)
	}
	authHandler.RegisterRoutes(api)
	handlers.NewBuildingHandler(db).RegisterRoutes(api)
	handlers.NewEquipmentGroupHandler(db).RegisterRoutes(api)
	handlers.NewEquipmentHandler(db).RegisterRoutes(api)
	handlers.NewRequestHandler(db).RegisterRoutes(api)
	handlers.NewRequestedEquipmentHandler(db).RegisterRoutes(api)

	r.Run("0.0.0.0:8001")
}

// backfillEquipmentBuildingID assigns a fallback building to any pre-existing equipment rows
// with a NULL building_id, before AutoMigrate enforces the NOT NULL constraint on that column.
// A fresh database has no equipment table yet, and one with buildings but no orphaned equipment
// has nothing to do either; both are no-ops.
func backfillEquipmentBuildingID(db *gorm.DB) error {
	if !db.Migrator().HasTable("equipment") || !db.Migrator().HasColumn("equipment", "building_id") {
		return nil
	}

	var orphaned int64
	if err := db.Table("equipment").Where("building_id IS NULL").Count(&orphaned).Error; err != nil {
		return err
	}
	if orphaned == 0 {
		return nil
	}

	var fallbackBuildingID uint
	if err := db.Table("buildings").Order("id").Limit(1).Pluck("id", &fallbackBuildingID).Error; err != nil {
		return err
	}
	if fallbackBuildingID == 0 {
		return fmt.Errorf("%d equipment row(s) have a NULL building_id, but no building exists to backfill them into; create a building first", orphaned)
	}

	return db.Exec("UPDATE equipment SET building_id = ? WHERE building_id IS NULL", fallbackBuildingID).Error
}
