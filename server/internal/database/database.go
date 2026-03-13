package database

import (
	"log"
	"os"

	"com.hermes.platform/internal/config"
	"com.hermes.platform/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func InitDB(cfg *config.Config) (*gorm.DB, error) {
	dsn := cfg.Database.GetDSN()

	gormLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			LogLevel: logger.Warn,
			Colorful: true,
		},
	)

	if cfg.Env == "development" {
		gormLogger = logger.Default.LogMode(logger.Info)
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormLogger,
	})
	if err != nil {
		return nil, err
	}
	return db, nil
}

func AutoMigrate(db *gorm.DB) {
	db.Exec(`
		DO $$ 
		BEGIN
			IF EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'uni_users_email') THEN
				ALTER TABLE users DROP CONSTRAINT uni_users_email;
			END IF;
		END $$;
	`)
	db.AutoMigrate(
		&models.User{},
		&models.Role{},
		&models.Permission{},
		&models.TestTask{},
		&models.TestDetail{},
		&models.TestRecord{},
		&models.APIToken{},
	)
}
