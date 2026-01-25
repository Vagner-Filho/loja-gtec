package products

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/lib/pq"
)

type Product struct {
	ID             int     `json:"id"`
	ProductID      int     `json:"productId"`
	Name           string  `json:"name"`
	Price          float64 `json:"price"`
	Image          string  `json:"image"`
	Category       string  `json:"category"`
	IsAvailable    bool    `json:"isAvailable"`
	BrandIDs       []int   `json:"brandIds,omitempty"`
	FitsProductIDs []int   `json:"fitsProductIds,omitempty"`
}

type Brand struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type ProductOption struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
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
	var productID int
	err := db.QueryRow("SELECT products.id, items.id, items.name, items.price, items.image, products.category, items.is_available FROM products JOIN items ON products.item_id = items.id WHERE items.id = $1", id).
		Scan(&productID, &p.ID, &p.Name, &p.Price, &p.Image, &p.Category, &p.IsAvailable)
	if err != nil {
		return nil, err
	}

	p.ProductID = productID
	brandIDs, err := getBrandIDsByProductID(productID)
	if err != nil {
		return nil, err
	}
	fitIDs, err := getFitsProductIDsByProductID(productID)
	if err != nil {
		return nil, err
	}

	p.BrandIDs = brandIDs
	p.FitsProductIDs = fitIDs
	return &p, nil
}

// CreateProduct creates a new product in the database
func CreateProduct(name string, price float64, image, category string, isAvailable bool, brandIDs, fitsProductIDs []int) (*Product, error) {
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

	var productID int
	err = tx.QueryRow(
		"INSERT INTO products (category, item_id) VALUES ($1, $2) RETURNING id",
		category, itemID,
	).Scan(&productID)
	if err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	if err := insertProductBrands(tx, productID, brandIDs); err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	if err := insertProductCompatibility(tx, productID, category, fitsProductIDs); err != nil {
		_ = tx.Rollback()
		return nil, err
	}

	if err = tx.Commit(); err != nil {
		return nil, err
	}

	return &Product{
		ID:             itemID,
		ProductID:      productID,
		Name:           name,
		Price:          price,
		Image:          image,
		Category:       category,
		IsAvailable:    isAvailable,
		BrandIDs:       uniqueIDs(brandIDs),
		FitsProductIDs: uniqueIDs(fitsProductIDs),
	}, nil
}

// UpdateProduct updates an existing product in the database
func UpdateProduct(id int, name string, price float64, image, category string, isAvailable bool, brandIDs, fitsProductIDs []int) error {
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

	var productID int
	if err := tx.QueryRow("SELECT id FROM products WHERE item_id = $1", id).Scan(&productID); err != nil {
		_ = tx.Rollback()
		return err
	}

	if _, err := tx.Exec("DELETE FROM product_brands WHERE product_id = $1", productID); err != nil {
		_ = tx.Rollback()
		return err
	}
	if err := insertProductBrands(tx, productID, brandIDs); err != nil {
		_ = tx.Rollback()
		return err
	}

	if _, err := tx.Exec("DELETE FROM product_compatibility WHERE part_product_id = $1", productID); err != nil {
		_ = tx.Rollback()
		return err
	}
	if err := insertProductCompatibility(tx, productID, category, fitsProductIDs); err != nil {
		_ = tx.Rollback()
		return err
	}

	if err = tx.Commit(); err != nil {
		return err
	}

	return nil
}

func GetAllBrands() ([]Brand, error) {
	rows, err := db.Query("SELECT id, name FROM brands ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var brands []Brand
	for rows.Next() {
		var b Brand
		if err := rows.Scan(&b.ID, &b.Name); err != nil {
			return nil, err
		}
		brands = append(brands, b)
	}

	return brands, nil
}

func CreateBrand(name string) (*Brand, error) {
	trimmed := strings.TrimSpace(name)
	if trimmed == "" {
		return nil, fmt.Errorf("nome da marca nao pode ser vazio")
	}

	var brand Brand
	brand.Name = trimmed
	err := db.QueryRow("INSERT INTO brands (name) VALUES ($1) RETURNING id", trimmed).Scan(&brand.ID)
	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" {
			return nil, fmt.Errorf("marca ja cadastrada")
		}
		return nil, err
	}

	return &brand, nil
}

func GetAllProductOptions() ([]ProductOption, error) {
	rows, err := db.Query("SELECT products.id, items.name FROM products JOIN items ON products.item_id = items.id ORDER BY items.name")
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

func getBrandIDsByProductID(productID int) ([]int, error) {
	rows, err := db.Query("SELECT brand_id FROM product_brands WHERE product_id = $1 ORDER BY brand_id", productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}

	return ids, nil
}

func getFitsProductIDsByProductID(productID int) ([]int, error) {
	rows, err := db.Query("SELECT fits_product_id FROM product_compatibility WHERE part_product_id = $1 ORDER BY fits_product_id", productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}

	return ids, nil
}

func insertProductBrands(tx *sql.Tx, productID int, brandIDs []int) error {
	for _, id := range uniqueIDs(brandIDs) {
		if id == 0 {
			continue
		}
		if _, err := tx.Exec("INSERT INTO product_brands (product_id, brand_id) VALUES ($1, $2)", productID, id); err != nil {
			return err
		}
	}
	return nil
}

func insertProductCompatibility(tx *sql.Tx, productID int, category string, fitsProductIDs []int) error {
	if len(fitsProductIDs) == 0 {
		return nil
	}
	if !isPartsCategory(category) {
		return fmt.Errorf("apenas produtos das categorias refis ou pecas podem ter compatibilidade")
	}
	for _, id := range uniqueIDs(fitsProductIDs) {
		if id == 0 {
			continue
		}
		if id == productID {
			return fmt.Errorf("produto nao pode ser compativel consigo mesmo")
		}
		if _, err := tx.Exec("INSERT INTO product_compatibility (part_product_id, fits_product_id) VALUES ($1, $2)", productID, id); err != nil {
			return err
		}
	}
	return nil
}

func isPartsCategory(category string) bool {
	return category == "refis" || category == "pecas"
}

func uniqueIDs(ids []int) []int {
	seen := make(map[int]struct{}, len(ids))
	unique := make([]int, 0, len(ids))
	for _, id := range ids {
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		unique = append(unique, id)
	}
	return unique
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
