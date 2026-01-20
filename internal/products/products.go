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
	rows, err := db.Query("SELECT items.id, items.name, items.price, items.image, products.category, items.is_available FROM products JOIN items ON products.item_id = items.id ORDER BY items.id DESC")
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

	rows, err = db.Query("SELECT items.id, items.name, items.price, items.image, products.category, items.is_available FROM products JOIN items ON products.item_id = items.id WHERE products.category = $1 ORDER BY items.id DESC", category)
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
	err := db.QueryRow("SELECT items.id, items.name, items.price, items.image, products.category, items.is_available FROM products JOIN items ON products.item_id = items.id WHERE items.id = $1", id).
		Scan(&p.ID, &p.Name, &p.Price, &p.Image, &p.Category, &p.IsAvailable)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

// CreateProduct creates a new product in the database
func CreateProduct(name string, price float64, image, category string, isAvailable bool) (*Product, error) {
	tx, err := db.Begin()
	if err != nil {
		return nil, err
	}

	var itemID int
	err = tx.QueryRow(
		"INSERT INTO items (name, price, image, is_available) VALUES ($1, $2, $3, $4) RETURNING id",
		name, price, image, isAvailable,
	).Scan(&itemID)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	_, err = tx.Exec(
		"INSERT INTO products (category, item_id) VALUES ($1, $2)",
		category, itemID,
	)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return &Product{
		ID:          itemID,
		Name:        name,
		Price:       price,
		Image:       image,
		Category:    category,
		IsAvailable: isAvailable,
	}, nil
}

// UpdateProduct updates an existing product in the database
func UpdateProduct(id int, name string, price float64, image, category string, isAvailable bool) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}

	result, err := tx.Exec(
		"UPDATE items SET name = $1, price = $2, image = $3, is_available = $4, updated_at = CURRENT_TIMESTAMP WHERE id = $5",
		name, price, image, isAvailable, id,
	)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	if rowsAffected == 0 {
		_ = tx.Rollback()
		return fmt.Errorf("product with id %d not found", id)
	}

	_, err = tx.Exec(
		"UPDATE products SET category = $1 WHERE item_id = $2",
		category, id,
	)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

// DeleteProduct deletes a product from the database
func DeleteProduct(id int) error {
	if id == 1 {
		return fmt.Errorf("Não é possível excluir o serviço de instalação")
	}

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	result, err := tx.Exec("DELETE FROM products WHERE item_id = $1", id)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	if rowsAffected == 0 {
		_ = tx.Rollback()
		return fmt.Errorf("product with id %d not found", id)
	}

	_, err = tx.Exec("DELETE FROM items WHERE id = $1", id)
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}
