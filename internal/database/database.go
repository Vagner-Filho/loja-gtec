package database

import (
	"database/sql"
	"fmt"
	"log"
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

	// Always run migrations first (safe in all environments).
	if err := RunMigrations(db); err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	// In non-development environments, never auto-apply schema.sql or seed.sql.
	if config.Environment != "development" {
		fmt.Println("Database migrations applied successfully!")
		return nil
	}

	// Development-only: if the database appears completely empty (no admin_users
	// table), apply the baseline schema and seed data. schema.sql is idempotent,
	// and seed.sql is strictly optional dev data.
	var tableExists bool
	err := db.QueryRow(`
		SELECT EXISTS (
			SELECT FROM information_schema.tables
			WHERE table_schema = 'public'
			AND table_name = 'admin_users'
		)`).Scan(&tableExists)
	if err != nil {
		return fmt.Errorf("failed to check if admin_users table exists: %w", err)
	}

	if !tableExists {
		schema, err := os.ReadFile("scripts/schema/schema.sql")
		if err != nil {
			return fmt.Errorf("failed to read schema file: %w", err)
		}
		if _, err := db.Exec(string(schema)); err != nil {
			return fmt.Errorf("failed to execute schema: %w", err)
		}
		fmt.Println("Development schema applied successfully!")

		seed, err := os.ReadFile("scripts/schema/seed.sql")
		if err == nil {
			if _, err := db.Exec(string(seed)); err != nil {
				return fmt.Errorf("failed to execute seed: %w", err)
			}
			fmt.Println("Development seed applied successfully!")
		} else if !os.IsNotExist(err) {
			return fmt.Errorf("failed to read seed file: %w", err)
		}
	}

	// Warn if no admin users exist (applies to both fresh and existing dev DBs).
	var adminCount int
	if err := db.QueryRow("SELECT COUNT(*) FROM admin_users").Scan(&adminCount); err != nil {
		return fmt.Errorf("failed to count admin users: %w", err)
	}
	if adminCount == 0 {
		log.Println("WARNING: No admin users found. Create one manually or use the admin setup utility.")
	}

	fmt.Println("Database migrations applied successfully!")
	return nil
}
