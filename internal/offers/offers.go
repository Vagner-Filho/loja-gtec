package offers

import (
	"database/sql"
	"fmt"
	"time"
)

// Offer represents a product that is on offer/promotion
type Offer struct {
	ID         int        `json:"id"`
	ProductID  int        `json:"productId"`
	Name       string     `json:"name"`
	Price      float64    `json:"price"`
	OfferPrice float64    `json:"offerPrice"`
	Category   string     `json:"category"`
	StartDate  *time.Time `json:"startDate,omitempty"`
	EndDate    *time.Time `json:"endDate,omitempty"`
	IsActive   bool       `json:"isActive"`
}

// OfferForm represents the form data for creating/updating an offer
type OfferForm struct {
	ProductID  int        `json:"productId"`
	OfferPrice float64    `json:"offerPrice"`
	StartDate  *time.Time `json:"startDate,omitempty"`
	EndDate    *time.Time `json:"endDate,omitempty"`
}

// ProductOption represents a product that can be selected for an offer
type ProductOption struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

var db *sql.DB

// SetDatabase sets the database connection for the offers package
func SetDatabase(database *sql.DB) {
	db = database
}

// IsOfferActive checks if an offer is currently active based on its dates and status
func IsOfferActive(offer Offer) bool {
	if !offer.IsActive {
		return false
	}
	now := time.Now()
	if offer.StartDate != nil && now.Before(*offer.StartDate) {
		return false
	}
	if offer.EndDate != nil && now.After(*offer.EndDate) {
		return false
	}
	return true
}

// GetActiveOffers retrieves all offers currently active for public display
func GetActiveOffers() ([]Offer, error) {
	query := `
		SELECT o.id, o.product_id, i.name, i.price, o.offer_price, 
		       p.category, o.start_date, o.end_date, o.is_active
		FROM offers o
		JOIN products p ON o.product_id = p.id
		JOIN items i ON p.item_id = i.id
		WHERE o.is_active = TRUE
		ORDER BY o.id DESC
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var offers []Offer
	now := time.Now()
	for rows.Next() {
		var o Offer
		var startDate, endDate sql.NullTime
		err := rows.Scan(&o.ID, &o.ProductID, &o.Name, &o.Price, &o.OfferPrice,
			&o.Category, &startDate, &endDate, &o.IsActive)
		if err != nil {
			return nil, err
		}

		if startDate.Valid {
			o.StartDate = &startDate.Time
		}
		if endDate.Valid {
			o.EndDate = &endDate.Time
		}

		// Only include if currently active based on dates
		if o.IsActive {
			if o.StartDate != nil && now.Before(*o.StartDate) {
				continue
			}
			if o.EndDate != nil && now.After(*o.EndDate) {
				continue
			}
			offers = append(offers, o)
		}
	}

	return offers, nil
}

// GetAllOffers retrieves all offers including inactive ones (for admin)
func GetAllOffers() ([]Offer, error) {
	query := `
		SELECT o.id, o.product_id, i.name, i.price, o.offer_price, 
		       p.category, o.start_date, o.end_date, o.is_active
		FROM offers o
		JOIN products p ON o.product_id = p.id
		JOIN items i ON p.item_id = i.id
		ORDER BY o.id DESC
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var offers []Offer
	for rows.Next() {
		var o Offer
		var startDate, endDate sql.NullTime
		err := rows.Scan(&o.ID, &o.ProductID, &o.Name, &o.Price, &o.OfferPrice,
			&o.Category, &startDate, &endDate, &o.IsActive)
		if err != nil {
			return nil, err
		}

		if startDate.Valid {
			o.StartDate = &startDate.Time
		}
		if endDate.Valid {
			o.EndDate = &endDate.Time
		}

		offers = append(offers, o)
	}

	return offers, nil
}

// GetProductsForOfferSelection retrieves products not currently on offer
func GetProductsForOfferSelection() ([]ProductOption, error) {
	query := `
		SELECT p.id, i.name 
		FROM products p
		JOIN items i ON p.item_id = i.id 
		WHERE NOT EXISTS (
			SELECT 1 FROM offers o WHERE o.product_id = p.id AND o.is_active = TRUE
		)
		ORDER BY i.name
	`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []ProductOption
	for rows.Next() {
		var p ProductOption
		if err := rows.Scan(&p.ID, &p.Name); err != nil {
			return nil, err
		}
		products = append(products, p)
	}

	return products, nil
}

// CreateOffer creates a new offer for a product
func CreateOffer(form OfferForm) error {
	// Validate offer price
	if form.OfferPrice <= 0 {
		return fmt.Errorf("preço de oferta deve ser maior que zero")
	}

	// Validate dates if both are provided
	if form.StartDate != nil && form.EndDate != nil {
		if form.EndDate.Before(*form.StartDate) {
			return fmt.Errorf("data de término deve ser posterior à data de início")
		}
	}

	// Check if offer already exists for this product
	var existingID int
	err := db.QueryRow("SELECT id FROM offers WHERE product_id = $1", form.ProductID).Scan(&existingID)
	if err == nil {
		// Offer exists, update it
		_, err = db.Exec(
			`UPDATE offers 
			 SET offer_price = $1, start_date = $2, end_date = $3, is_active = TRUE, updated_at = CURRENT_TIMESTAMP
			 WHERE product_id = $4`,
			form.OfferPrice, form.StartDate, form.EndDate, form.ProductID,
		)
		if err != nil {
			return err
		}
		return nil
	}

	// Create new offer
	_, err = db.Exec(
		`INSERT INTO offers (product_id, offer_price, start_date, end_date, is_active)
		 VALUES ($1, $2, $3, $4, TRUE)`,
		form.ProductID, form.OfferPrice, form.StartDate, form.EndDate,
	)
	if err != nil {
		return err
	}

	return nil
}

// UpdateOffer updates an existing offer
func UpdateOffer(offerID int, form OfferForm) error {
	// Validate offer price
	if form.OfferPrice <= 0 {
		return fmt.Errorf("preço de oferta deve ser maior que zero")
	}

	// Validate dates if both are provided
	if form.StartDate != nil && form.EndDate != nil {
		if form.EndDate.Before(*form.StartDate) {
			return fmt.Errorf("data de término deve ser posterior à data de início")
		}
	}

	result, err := db.Exec(
		`UPDATE offers 
		 SET offer_price = $1, start_date = $2, end_date = $3, updated_at = CURRENT_TIMESTAMP
		 WHERE id = $4`,
		form.OfferPrice, form.StartDate, form.EndDate, offerID,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("oferta não encontrada")
	}

	return nil
}

// ToggleOfferStatus toggles the active status of an offer (soft delete)
func ToggleOfferStatus(offerID int) (bool, error) {
	query := `
		UPDATE offers 
		SET is_active = NOT is_active, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 
		RETURNING is_active
	`

	var isActive bool
	err := db.QueryRow(query, offerID).Scan(&isActive)
	if err != nil {
		return false, err
	}

	return isActive, nil
}

// GetOfferByProductID retrieves a single offer by product ID
func GetOfferByProductID(productID int) (*Offer, error) {
	query := `
		SELECT o.id, o.product_id, i.name, i.price, o.offer_price, 
		       p.category, o.start_date, o.end_date, o.is_active
		FROM offers o
		JOIN products p ON o.product_id = p.id
		JOIN items i ON p.item_id = i.id
		WHERE o.product_id = $1
	`

	var o Offer
	var startDate, endDate sql.NullTime
	err := db.QueryRow(query, productID).Scan(
		&o.ID, &o.ProductID, &o.Name, &o.Price, &o.OfferPrice,
		&o.Category, &startDate, &endDate, &o.IsActive,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("oferta não encontrada")
		}
		return nil, err
	}

	if startDate.Valid {
		o.StartDate = &startDate.Time
	}
	if endDate.Valid {
		o.EndDate = &endDate.Time
	}

	return &o, nil
}

// GetActiveOfferByProductID retrieves an active offer for a product if one exists
func GetActiveOfferByProductID(productID int) (*Offer, error) {
	query := `
		SELECT o.id, o.product_id, i.name, i.price, o.offer_price, 
		       p.category, o.start_date, o.end_date, o.is_active
		FROM offers o
		JOIN products p ON o.product_id = p.id
		JOIN items i ON p.item_id = i.id
		WHERE o.product_id = $1 AND o.is_active = TRUE
	`

	var o Offer
	var startDate, endDate sql.NullTime
	err := db.QueryRow(query, productID).Scan(
		&o.ID, &o.ProductID, &o.Name, &o.Price, &o.OfferPrice,
		&o.Category, &startDate, &endDate, &o.IsActive,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	if startDate.Valid {
		o.StartDate = &startDate.Time
	}
	if endDate.Valid {
		o.EndDate = &endDate.Time
	}

	// Check if offer is actually active based on dates
	if !IsOfferActive(o) {
		return nil, nil
	}

	return &o, nil
}
