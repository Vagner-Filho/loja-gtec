package banners

import (
	"database/sql"
	"fmt"
)

// Banner represents a promotional banner
type Banner struct {
	ID           int    `json:"id"`
	ImagePath    string `json:"image_path"`
	Title        string `json:"title"`
	LinkURL      string `json:"link_url"`
	DisplayOrder int    `json:"display_order"`
	IsActive     bool   `json:"is_active"`
	CreatedAt    string `json:"created_at"`
}

var db *sql.DB

// SetDatabase sets the database connection for the banners package
func SetDatabase(database *sql.DB) {
	db = database
}

// GetAllBanners retrieves all banners from the database
func GetAllBanners() ([]Banner, error) {
	rows, err := db.Query("SELECT id, image_path, title, link_url, display_order, is_active, created_at FROM banners ORDER BY display_order ASC, id ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var banners []Banner
	for rows.Next() {
		var b Banner
		if err := rows.Scan(&b.ID, &b.ImagePath, &b.Title, &b.LinkURL, &b.DisplayOrder, &b.IsActive, &b.CreatedAt); err != nil {
			return nil, err
		}
		banners = append(banners, b)
	}

	return banners, nil
}

// GetActiveBanners retrieves only active banners ordered by display_order
func GetActiveBanners() ([]Banner, error) {
	rows, err := db.Query("SELECT id, image_path, title, link_url, display_order, is_active, created_at FROM banners WHERE is_active = TRUE ORDER BY display_order ASC, id ASC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var banners []Banner
	for rows.Next() {
		var b Banner
		if err := rows.Scan(&b.ID, &b.ImagePath, &b.Title, &b.LinkURL, &b.DisplayOrder, &b.IsActive, &b.CreatedAt); err != nil {
			return nil, err
		}
		banners = append(banners, b)
	}

	return banners, nil
}

// CreateBanner creates a new banner
func CreateBanner(imagePath, title, linkURL string) (*Banner, error) {
	if title == "" {
		return nil, fmt.Errorf("title is required")
	}

	query := `
		INSERT INTO banners (image_path, title, link_url, display_order)
		VALUES ($1, $2, $3, (SELECT COALESCE(MAX(display_order), 0) + 1 FROM banners))
		RETURNING id, image_path, title, link_url, display_order, is_active, created_at
	`

	var b Banner
	err := db.QueryRow(query, imagePath, title, linkURL).Scan(
		&b.ID, &b.ImagePath, &b.Title, &b.LinkURL, &b.DisplayOrder, &b.IsActive, &b.CreatedAt,
	)
	if err != nil {
		return nil, err
	}

	return &b, nil
}

// DeleteBanner deletes a banner by ID
func DeleteBanner(id int) error {
	result, err := db.Exec("DELETE FROM banners WHERE id = $1", id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("banner not found")
	}

	return nil
}

// ToggleBannerStatus toggles the active status of a banner
func ToggleBannerStatus(id int) (bool, error) {
	query := `
		UPDATE banners 
		SET is_active = NOT is_active 
		WHERE id = $1 
		RETURNING is_active
	`

	var isActive bool
	err := db.QueryRow(query, id).Scan(&isActive)
	if err != nil {
		return false, err
	}

	return isActive, nil
}

// UpdateBannerOrder updates the display order of a banner
func UpdateBannerOrder(id, newOrder int) error {
	result, err := db.Exec("UPDATE banners SET display_order = $1 WHERE id = $2", newOrder, id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("banner not found")
	}

	return nil
}

// GetBannerByID retrieves a single banner by ID
func GetBannerByID(id int) (*Banner, error) {
	var b Banner
	err := db.QueryRow(
		"SELECT id, image_path, title, link_url, display_order, is_active, created_at FROM banners WHERE id = $1",
		id,
	).Scan(&b.ID, &b.ImagePath, &b.Title, &b.LinkURL, &b.DisplayOrder, &b.IsActive, &b.CreatedAt)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("banner not found")
		}
		return nil, err
	}

	return &b, nil
}
