package database

import (
	"database/sql"
	"fmt"
	"lojagtec/internal/admin"
	"os"

	"github.com/BurntSushi/toml"
	_ "github.com/lib/pq"
)

type Config struct {
	Environment string `toml:"environment"`
	Database    struct {
		Host     string `toml:"host"`
		Port     int    `toml:"port"`
		User     string `toml:"user"`
		Password string `toml:"password"`
		DBName   string `toml:"dbname"`
		SSLMode  string `toml:"sslmode"`
	}
}

func Connect() (*sql.DB, error) {
	var config Config
	if _, err := toml.DecodeFile("configs/config.toml", &config); err != nil {
		return nil, err
	}

	psqlInfo := fmt.Sprintf("host=%s port=%d user=%s "+
		"password=%s dbname=%s sslmode=%s",
		config.Database.Host, config.Database.Port, config.Database.User,
		config.Database.Password, config.Database.DBName, config.Database.SSLMode)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	fmt.Println("Successfully connected to the database!")
	return db, nil
}

func RunSchema(db *sql.DB) error {
	var config Config
	if _, err := toml.DecodeFile("configs/config.toml", &config); err != nil {
		return err
	}

	schema, err := os.ReadFile("scripts/schema/schema.sql")

	if err != nil {
		return fmt.Errorf("failed to read schema file: %w", err)
	}

	_, err = db.Exec(string(schema))

	if err != nil {
		return fmt.Errorf("failed to execute schema: %w", err)
	}

	var adminCount uint8
	err = db.QueryRow("SELECT COUNT(*) FROM admin_users").Scan(&adminCount)
	if err != nil {
		return fmt.Errorf("failed to create admin: %w", err)
	}
	if adminCount == 0 && config.Environment == "development" {
		admin.CreateAdmin("admin", "123")
		admin.CreateAdminWithRole("p_admin", "123", "product_admin")
		seed, err := os.ReadFile("scripts/schema/seed.sql")

		if err != nil {
			return fmt.Errorf("failed to read seed file: %w", err)
		}

		_, err = db.Exec(string(seed))

		if err != nil {
			return fmt.Errorf("failed to execute seed: %w", err)
		}
		fmt.Println("Database seed applied successfully!")
	}

	fmt.Println("Database schema applied successfully!")
	return nil
}
