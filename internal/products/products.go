package products

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/lib/pq"
)

type Product struct {
	ID             int        `json:"id"`
	ProductID      int        `json:"productId"`
	Name           string     `json:"name"`
	Price          float64    `json:"price"`
	Image          string     `json:"image"`
	Category       string     `json:"category"`
	IsAvailable    bool       `json:"isAvailable"`
	IsOnOffer      bool       `json:"isOnOffer"`
	OfferPrice     float64    `json:"offerPrice,omitempty"`
	OfferStartDate *time.Time `json:"offerStartDate,omitempty"`
	OfferEndDate   *time.Time `json:"offerEndDate,omitempty"`
	BrandIDs       []int      `json:"brandIds,omitempty"`
	FitsProductIDs []int      `json:"fitsProductIds,omitempty"`
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

// scanProduct scans a product row with optional offer data
func scanProduct(rows *sql.Rows) (Product, error) {
	var p Product
	var offerID sql.NullInt64
	var offerPrice sql.NullFloat64
	var startDate, endDate sql.NullTime
	var isActive sql.NullBool

	err := rows.Scan(&p.ID, &p.ProductID, &p.Name, &p.Price, &p.Image, &p.Category, &p.IsAvailable,
		&offerID, &offerPrice, &startDate, &endDate, &isActive)
	if err != nil {
		return p, err
	}

	// Check if product has an active offer
	if offerID.Valid && isActive.Valid && isActive.Bool {
		now := time.Now()
		offerStart := startDate.Time
		offerEnd := endDate.Time

		// Check if offer is currently active based on dates
		isCurrentlyActive := true
		if startDate.Valid && now.Before(offerStart) {
			isCurrentlyActive = false
		}
		if endDate.Valid && now.After(offerEnd) {
			isCurrentlyActive = false
		}

		if isCurrentlyActive {
			p.IsOnOffer = true
			if offerPrice.Valid {
				p.OfferPrice = offerPrice.Float64
			}
			if startDate.Valid {
				p.OfferStartDate = &offerStart
			}
			if endDate.Valid {
				p.OfferEndDate = &offerEnd
			}
		}
	}

	return p, nil
}

// scanSearchProduct scans a product row with similarity score from search queries
func scanSearchProduct(rows *sql.Rows) (Product, error) {
	var p Product
	var offerID sql.NullInt64
	var offerPrice sql.NullFloat64
	var startDate, endDate sql.NullTime
	var isActive sql.NullBool
	var similarityScore float64 // Ignored, just for ORDER BY

	err := rows.Scan(&p.ID, &p.ProductID, &p.Name, &p.Price, &p.Image, &p.Category, &p.IsAvailable,
		&offerID, &offerPrice, &startDate, &endDate, &isActive, &similarityScore)
	if err != nil {
		return p, err
	}

	// Check if product has an active offer
	if offerID.Valid && isActive.Valid && isActive.Bool {
		now := time.Now()
		offerStart := startDate.Time
		offerEnd := endDate.Time

		// Check if offer is currently active based on dates
		isCurrentlyActive := true
		if startDate.Valid && now.Before(offerStart) {
			isCurrentlyActive = false
		}
		if endDate.Valid && now.After(offerEnd) {
			isCurrentlyActive = false
		}

		if isCurrentlyActive {
			p.IsOnOffer = true
			if offerPrice.Valid {
				p.OfferPrice = offerPrice.Float64
			}
			if startDate.Valid {
				p.OfferStartDate = &offerStart
			}
			if endDate.Valid {
				p.OfferEndDate = &offerEnd
			}
		}
	}

	return p, nil
}

// GetAllProducts retrieves all products from the database
func GetAllProducts() ([]Product, error) {
	query := `SELECT items.id, products.id, items.name, items.price, items.image, products.category, 
		items.is_available, o.id, o.offer_price, o.start_date, o.end_date, o.is_active
		FROM products 
		JOIN items ON products.item_id = items.id 
		LEFT JOIN offers o ON products.id = o.product_id AND o.is_active = TRUE
		ORDER BY items.id DESC`

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		p, err := scanProduct(rows)
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}

	return products, nil
}

// GetProductsByCategory retrieves products by category from the database
func GetProductsByCategory(category string) ([]Product, error) {
	return GetProductsByCategoryAndBrands(category, nil)
}

// GetProductsByCategoryAndBrands retrieves products by category and brand filters from the database
func GetProductsByCategoryAndBrands(category string, brandIDs []int) ([]Product, error) {
	var rows *sql.Rows
	var err error

	if category == "" && len(brandIDs) == 0 {
		return GetAllProducts()
	}

	query := `SELECT DISTINCT items.id, products.id, items.name, items.price, items.image, products.category, 
		items.is_available, o.id, o.offer_price, o.start_date, o.end_date, o.is_active
		FROM products 
		JOIN items ON products.item_id = items.id
		LEFT JOIN offers o ON products.id = o.product_id AND o.is_active = TRUE`
	var args []interface{}
	var conditions []string

	if len(brandIDs) > 0 {
		query += " JOIN product_brands ON products.id = product_brands.product_id"
		conditions = append(conditions, "product_brands.brand_id = ANY($"+string(rune('0'+len(args)+1))+")")
		args = append(args, pq.Array(brandIDs))
	}

	if category != "" {
		conditions = append(conditions, "products.category = $"+string(rune('0'+len(args)+1)))
		args = append(args, category)
	}

	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}

	query += " ORDER BY items.id DESC"

	rows, err = db.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		p, err := scanProduct(rows)
		if err != nil {
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
	var offerID sql.NullInt64
	var offerPrice sql.NullFloat64
	var startDate, endDate sql.NullTime
	var isActive sql.NullBool

	query := `SELECT items.id, products.id, items.name, items.price, items.image, 
		products.category, items.is_available, o.id, o.offer_price, 
		o.start_date, o.end_date, o.is_active
		FROM products 
		JOIN items ON products.item_id = items.id 
		LEFT JOIN offers o ON products.id = o.product_id AND o.is_active = TRUE
		WHERE items.id = $1`

	err := db.QueryRow(query, id).Scan(
		&p.ID, &productID, &p.Name, &p.Price, &p.Image, &p.Category, &p.IsAvailable,
		&offerID, &offerPrice, &startDate, &endDate, &isActive,
	)
	if err != nil {
		return nil, err
	}

	// Check if product has an active offer
	if offerID.Valid && isActive.Valid && isActive.Bool {
		now := time.Now()
		offerStart := startDate.Time
		offerEnd := endDate.Time

		// Check if offer is currently active based on dates
		isCurrentlyActive := true
		if startDate.Valid && now.Before(offerStart) {
			isCurrentlyActive = false
		}
		if endDate.Valid && now.After(offerEnd) {
			isCurrentlyActive = false
		}

		if isCurrentlyActive {
			p.IsOnOffer = true
			p.ProductID = productID
			if offerPrice.Valid {
				p.OfferPrice = offerPrice.Float64
			}
			if startDate.Valid {
				p.OfferStartDate = &offerStart
			}
			if endDate.Valid {
				p.OfferEndDate = &offerEnd
			}
		}
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

// GetCurrentPrice returns the effective price (offer price if active, otherwise regular price)
func GetCurrentPrice(product Product) float64 {
	if product.IsOnOffer && product.OfferPrice > 0 {
		return product.OfferPrice
	}
	return product.Price
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

// GetBrandNamesByProductID returns brand names for a product
func GetBrandNamesByProductID(productID int) ([]Brand, error) {
	query := `SELECT b.id, b.name 
		FROM brands b 
		JOIN product_brands pb ON b.id = pb.brand_id 
		WHERE pb.product_id = $1 
		ORDER BY b.name`

	rows, err := db.Query(query, productID)
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

// GetCompatibleProductsByProductID returns products compatible with this part
func GetCompatibleProductsByProductID(productID int) ([]Product, error) {
	query := `SELECT items.id, products.id, items.name, items.price, items.image, products.category, 
		items.is_available, o.id, o.offer_price, o.start_date, o.end_date, o.is_active
		FROM product_compatibility pc
		JOIN products ON pc.fits_product_id = products.id
		JOIN items ON products.item_id = items.id
		LEFT JOIN offers o ON products.id = o.product_id AND o.is_active = TRUE
		WHERE pc.part_product_id = $1
		ORDER BY items.name`

	rows, err := db.Query(query, productID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		p, err := scanProduct(rows)
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}

	return products, nil
}

// BrandSearchResult represents a brand with its product count for search results
type BrandSearchResult struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	ProductCount int    `json:"productCount"`
}

// SearchProducts searches for products using fuzzy matching (trigram similarity)
func SearchProducts(query string, limit int) ([]Product, error) {
	if query == "" {
		return []Product{}, nil
	}

	// Use trigram similarity for fuzzy search with ILIKE as fallback
	// similarity > 0.3 provides good fuzzy matching while filtering out irrelevant results
	searchQuery := `%` + query + `%`

	rows, err := db.Query(`
		SELECT DISTINCT items.id, products.id, items.name, items.price, items.image, products.category, 
			items.is_available, o.id, o.offer_price, o.start_date, o.end_date, o.is_active,
			similarity(items.name, $2) as similarity_score
		FROM products 
		JOIN items ON products.item_id = items.id 
		LEFT JOIN offers o ON products.id = o.product_id AND o.is_active = TRUE
		WHERE items.name ILIKE $1 
			OR similarity(items.name, $2) > 0.3
		ORDER BY similarity_score DESC, items.name
		LIMIT $3`,
		searchQuery, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		p, err := scanSearchProduct(rows)
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}

	return products, nil
}

// SearchBrandsWithCount searches for brands using fuzzy matching and returns product counts
func SearchBrandsWithCount(query string, limit int) ([]BrandSearchResult, error) {
	if query == "" {
		return []BrandSearchResult{}, nil
	}

	// Use trigram similarity for fuzzy search with ILIKE as fallback
	searchQuery := `%` + query + `%`

	rows, err := db.Query(`
		SELECT b.id, b.name, COUNT(DISTINCT pb.product_id) as product_count
		FROM brands b
		LEFT JOIN product_brands pb ON b.id = pb.brand_id
		WHERE b.name ILIKE $1 
			OR similarity(b.name, $2) > 0.3
		GROUP BY b.id, b.name
		ORDER BY similarity(b.name, $2) DESC, b.name
		LIMIT $3`,
		searchQuery, query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var brands []BrandSearchResult
	for rows.Next() {
		var b BrandSearchResult
		if err := rows.Scan(&b.ID, &b.Name, &b.ProductCount); err != nil {
			return nil, err
		}
		brands = append(brands, b)
	}

	return brands, nil
}

// GetRelatedProducts returns products in the same category (excluding the given product)
func GetRelatedProducts(excludeProductID int, category string, limit int) ([]Product, error) {
	query := `SELECT items.id, products.id, items.name, items.price, items.image, products.category, 
		items.is_available, o.id, o.offer_price, o.start_date, o.end_date, o.is_active
		FROM products 
		JOIN items ON products.item_id = items.id 
		LEFT JOIN offers o ON products.id = o.product_id AND o.is_active = TRUE
		WHERE products.category = $1 AND products.id != $2 AND items.is_available = TRUE
		ORDER BY RANDOM()
		LIMIT $3`

	rows, err := db.Query(query, category, excludeProductID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var products []Product
	for rows.Next() {
		p, err := scanProduct(rows)
		if err != nil {
			return nil, err
		}
		products = append(products, p)
	}

	return products, nil
}
