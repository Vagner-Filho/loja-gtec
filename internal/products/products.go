package products

import (
	"database/sql"
	"fmt"
)

type Product struct {
	ID          int     `json:"id"`
	Name        string  `json:"name"`
	Price       float64 `json:"price"`
	Image       string  `json:"image"`
	Category    string  `json:"category"`
	IsAvailable bool    `json:"isAvailable"`
}

var db *sql.DB

// SetDatabase sets the database connection for the products package
func SetDatabase(database *sql.DB) {
	db = database
}

// GetAllProducts retrieves all products from the database
func GetAllProducts() ([]Product, error) {
	rows, err := db.Query("SELECT id, name, price, image, category, is_available FROM products ORDER BY id DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Price, &p.Image, &p.Category, &p.IsAvailable); err != nil {
			return nil, err
		}
		products = append(products, p)
	}

	return products, nil
}

// GetProductsByCategory retrieves products by category from the database
func GetProductsByCategory(category string) ([]Product, error) {
	var rows *sql.Rows
	var err error

	if category == "" {
		return GetAllProducts()
	}

	rows, err = db.Query("SELECT id, name, price, image, category, is_available FROM products WHERE category = $1 ORDER BY id DESC", category)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		var p Product
		if err := rows.Scan(&p.ID, &p.Name, &p.Price, &p.Image, &p.Category, &p.IsAvailable); err != nil {
			return nil, err
		}
		products = append(products, p)
	}

	return products, nil
}

// GetProductByID retrieves a single product by ID
func GetProductByID(id int) (*Product, error) {
	var p Product
	err := db.QueryRow("SELECT id, name, price, image, category, is_available FROM products WHERE id = $1", id).
		Scan(&p.ID, &p.Name, &p.Price, &p.Image, &p.Category, &p.IsAvailable)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// CreateProduct creates a new product in the database
func CreateProduct(name string, price float64, image, category string, isAvailable bool) (*Product, error) {
	var id int
	err := db.QueryRow(
		"INSERT INTO products (name, price, image, category, is_available) VALUES ($1, $2, $3, $4, $5) RETURNING id",
		name, price, image, category, isAvailable,
	).Scan(&id)

	if err != nil {
		return nil, err
	}

	return &Product{
		ID:          id,
		Name:        name,
		Price:       price,
		Image:       image,
		Category:    category,
		IsAvailable: isAvailable,
	}, nil
}

// UpdateProduct updates an existing product in the database
func UpdateProduct(id int, name string, price float64, image, category string, isAvailable bool) error {
	result, err := db.Exec(
		"UPDATE products SET name = $1, price = $2, image = $3, category = $4, is_available = $5, updated_at = CURRENT_TIMESTAMP WHERE id = $6",
		name, price, image, category, isAvailable, id,
	)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("product with id %d not found", id)
	}

	return nil
}

// DeleteProduct deletes a product from the database
func DeleteProduct(id int) error {
	result, err := db.Exec("DELETE FROM products WHERE id = $1", id)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return fmt.Errorf("product with id %d not found", id)
	}

	return nil
}
