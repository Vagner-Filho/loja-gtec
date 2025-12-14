package orders

import (
	"database/sql"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

// Order represents a customer order
type Order struct {
	ID              int       `json:"id"`
	OrderNumber     string    `json:"order_number"`
	Email           string    `json:"email"`
	Phone           string    `json:"phone"`
	FirstName       string    `json:"first_name"`
	LastName        string    `json:"last_name"`
	Address         string    `json:"address"`
	Neighborhood    string    `json:"neighborhood"`
	City            string    `json:"city"`
	State           string    `json:"state"`
	ZipCode         string    `json:"zip_code"`
	Apartment       string    `json:"apartment"`
	PaymentMethod   string    `json:"payment_method"`
	PaymentStatus   string    `json:"payment_status"`
	StripePaymentID string    `json:"stripe_payment_id"`
	TotalAmount     float64   `json:"total_amount"`
	Status          string    `json:"status"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// OrderItem represents an item in an order
type OrderItem struct {
	ID          int       `json:"id"`
	OrderID     int       `json:"order_id"`
	ProductID   int       `json:"product_id"`
	ProductName string    `json:"product_name"`
	Quantity    int       `json:"quantity"`
	UnitPrice   float64   `json:"unit_price"`
	TotalPrice  float64   `json:"total_price"`
	CreatedAt   time.Time `json:"created_at"`
}

// CheckoutForm represents the checkout form data
type CheckoutForm struct {
	Email         string `json:"email"`
	Phone         string `json:"phone"`
	FirstName     string `json:"first_name"`
	LastName      string `json:"last_name"`
	Address       string `json:"address"`
	Neighborhood  string `json:"neighborhood"`
	City          string `json:"city"`
	State         string `json:"state"`
	ZipCode       string `json:"zip_code"`
	Apartment     string `json:"apartment"`
	PaymentMethod string `json:"payment_method"`

	// Payment method specific fields
	CardName   string `json:"card_name,omitempty"`
	CardNumber string `json:"card_number,omitempty"`
	Expiry     string `json:"expiry,omitempty"`
	CVV        string `json:"cvv,omitempty"`
	CPF        string `json:"cpf,omitempty"`
	PixKey     string `json:"pix_key,omitempty"`

	// Cart items
	CartItems []CartItem `json:"cart_items"`
}

// CartItem represents a cart item
type CartItem struct {
	ID       int     `json:"id"`
	Name     string  `json:"name"`
	Price    float64 `json:"price"`
	Quantity int     `json:"quantity"`
}

// ValidationError represents a field validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationResult represents the result of form validation
type ValidationResult struct {
	IsValid bool              `json:"is_valid"`
	Errors  []ValidationError `json:"errors"`
}

var db *sql.DB

// SetDatabase sets the database connection for the orders package
func SetDatabase(database *sql.DB) {
	db = database
}

// ValidateEmail validates email format
func ValidateEmail(email string) *ValidationError {
	email = strings.TrimSpace(email)
	if email == "" {
		return &ValidationError{Field: "email", Message: "Email é obrigatório"}
	}

	emailRegex := regexp.MustCompile(`^[^\s@]+@[^\s@]+\.[^\s@]+$`)
	if !emailRegex.MatchString(email) {
		return &ValidationError{Field: "email", Message: "Por favor, insira um email válido"}
	}

	return nil
}

// ValidatePhone validates phone number
func ValidatePhone(phone string) *ValidationError {
	phone = strings.TrimSpace(phone)
	if phone == "" {
		return &ValidationError{Field: "phone", Message: "Telefone é obrigatório"}
	}

	// Remove non-digit characters
	cleaned := regexp.MustCompile(`\D`).ReplaceAllString(phone, "")
	if len(cleaned) < 10 {
		return &ValidationError{Field: "phone", Message: "Por favor, insira um telefone válido"}
	}

	return nil
}

// ValidateName validates first or last name
func ValidateName(name, field string) *ValidationError {
	name = strings.TrimSpace(name)
	if name == "" {
		if field == "firstName" {
			return &ValidationError{Field: "firstName", Message: "Nome é obrigatório"}
		}
		return &ValidationError{Field: "lastName", Message: "Sobrenome é obrigatório"}
	}

	if len(name) < 2 {
		if field == "firstName" {
			return &ValidationError{Field: "firstName", Message: "Nome deve ter pelo menos 2 caracteres"}
		}
		return &ValidationError{Field: "lastName", Message: "Sobrenome deve ter pelo menos 2 caracteres"}
	}

	return nil
}

// ValidateAddress validates address fields
func ValidateAddress(address, neighborhood, city, state, zipCode string) []ValidationError {
	var errors []ValidationError

	address = strings.TrimSpace(address)
	if address == "" {
		errors = append(errors, ValidationError{Field: "address", Message: "Endereço é obrigatório"})
	}

	neighborhood = strings.TrimSpace(neighborhood)
	if neighborhood == "" {
		errors = append(errors, ValidationError{Field: "neighborhood", Message: "Bairro é obrigatório"})
	}

	city = strings.TrimSpace(city)
	if city == "" {
		errors = append(errors, ValidationError{Field: "city", Message: "Cidade é obrigatória"})
	}

	state = strings.TrimSpace(state)
	if state == "" {
		errors = append(errors, ValidationError{Field: "state", Message: "Estado é obrigatório"})
	}

	// Validate Campo Grande, MS restriction
	if !ValidateCampoGrandeAddress(city, state) {
		errors = append(errors, ValidationError{Field: "city", Message: "Atendemos apenas clientes em Campo Grande, MS"})
	}

	zipCode = strings.TrimSpace(zipCode)
	if zipCode == "" {
		errors = append(errors, ValidationError{Field: "zipCode", Message: "CEP é obrigatório"})
	} else {
		// Validate CEP format (00000-000 or 00000000)
		cepRegex := regexp.MustCompile(`^\d{5}-?\d{3}$`)
		if !cepRegex.MatchString(zipCode) {
			errors = append(errors, ValidationError{Field: "zipCode", Message: "Por favor, insira um CEP válido"})
		}
	}

	return errors
}

// ValidateCampoGrandeAddress validates that address is in Campo Grande, MS
func ValidateCampoGrandeAddress(city, state string) bool {
	city = strings.ToLower(strings.TrimSpace(city))
	state = strings.ToUpper(strings.TrimSpace(state))
	return city == "campo grande" && state == "MS"
}

// ValidateCreditCard validates credit card information
func ValidateCreditCard(cardName, cardNumber, expiry, cvv string) []ValidationError {
	var errors []ValidationError

	cardName = strings.TrimSpace(cardName)
	if cardName == "" {
		errors = append(errors, ValidationError{Field: "cardName", Message: "Nome no cartão é obrigatório"})
	}

	cardNumber = strings.TrimSpace(cardNumber)
	if cardNumber == "" {
		errors = append(errors, ValidationError{Field: "cardNumber", Message: "Número do cartão é obrigatório"})
	} else {
		// Remove spaces and validate with Luhn algorithm
		cleaned := regexp.MustCompile(`\s`).ReplaceAllString(cardNumber, "")
		if !validateCardNumberLuhn(cleaned) {
			errors = append(errors, ValidationError{Field: "cardNumber", Message: "Por favor, insira um número de cartão válido"})
		}
	}

	expiry = strings.TrimSpace(expiry)
	if expiry == "" {
		errors = append(errors, ValidationError{Field: "expiry", Message: "Data de validade é obrigatória"})
	} else {
		if !validateExpiry(expiry) {
			errors = append(errors, ValidationError{Field: "expiry", Message: "Por favor, insira uma data de validade válida"})
		}
	}

	cvv = strings.TrimSpace(cvv)
	if cvv == "" {
		errors = append(errors, ValidationError{Field: "cvv", Message: "CVV é obrigatório"})
	} else {
		// Validate CVV (3-4 digits)
		cvvRegex := regexp.MustCompile(`^\d{3,4}$`)
		if !cvvRegex.MatchString(cvv) {
			errors = append(errors, ValidationError{Field: "cvv", Message: "Por favor, insira um CVV válido"})
		}
	}

	return errors
}

// validateCardNumberLuhn validates credit card number using Luhn algorithm
func validateCardNumberLuhn(cardNumber string) bool {
	// Remove non-digit characters
	cleaned := regexp.MustCompile(`\D`).ReplaceAllString(cardNumber, "")

	// Check length (13-19 digits for most cards)
	if len(cleaned) < 13 || len(cleaned) > 19 {
		return false
	}

	// Luhn algorithm
	sum := 0
	isEven := false

	for i := len(cleaned) - 1; i >= 0; i-- {
		digit := int(cleaned[i] - '0')

		if isEven {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
		isEven = !isEven
	}

	return sum%10 == 0
}

// validateExpiry validates card expiry date (MM/AA format)
func validateExpiry(expiry string) bool {
	// Check format (MM/AA)
	expiryRegex := regexp.MustCompile(`^\d{2}/\d{2}$`)
	if !expiryRegex.MatchString(expiry) {
		return false
	}

	parts := strings.Split(expiry, "/")
	if len(parts) != 2 {
		return false
	}

	month, err1 := strconv.Atoi(parts[0])
	year, err2 := strconv.Atoi("20" + parts[1]) // Convert YY to YYYY

	if err1 != nil || err2 != nil {
		return false
	}

	if month < 1 || month > 12 {
		return false
	}

	now := time.Now()
	currentYear := now.Year()
	currentMonth := int(now.Month())

	if year < currentYear {
		return false
	}
	if year == currentYear && month < currentMonth {
		return false
	}

	return true
}

// ValidateCPF validates CPF/CNPJ
func ValidateCPF(cpf string) *ValidationError {
	cpf = strings.TrimSpace(cpf)
	if cpf == "" {
		return &ValidationError{Field: "cpf", Message: "CPF/CNPJ é obrigatório"}
	}

	// Remove non-digit characters
	cleaned := regexp.MustCompile(`\D`).ReplaceAllString(cpf, "")

	if len(cleaned) != 11 && len(cleaned) != 14 {
		return &ValidationError{Field: "cpf", Message: "Por favor, insira um CPF ou CNPJ válido"}
	}

	return nil
}

// ValidatePixKey validates PIX key
func ValidatePixKey(pixKey string) *ValidationError {
	pixKey = strings.TrimSpace(pixKey)
	if pixKey == "" {
		return &ValidationError{Field: "pixKey", Message: "Chave PIX é obrigatória"}
	}

	if len(pixKey) < 5 {
		return &ValidationError{Field: "pixKey", Message: "Chave PIX inválida"}
	}

	return nil
}

// ValidateCheckoutForm validates the entire checkout form
func ValidateCheckoutForm(form CheckoutForm) ValidationResult {
	var errors []ValidationError

	// Validate email
	if err := ValidateEmail(form.Email); err != nil {
		errors = append(errors, *err)
	}

	// Validate phone
	if err := ValidatePhone(form.Phone); err != nil {
		errors = append(errors, *err)
	}

	// Validate names
	if err := ValidateName(form.FirstName, "firstName"); err != nil {
		errors = append(errors, *err)
	}
	if err := ValidateName(form.LastName, "lastName"); err != nil {
		errors = append(errors, *err)
	}

	// Validate address
	addressErrors := ValidateAddress(form.Address, form.Neighborhood, form.City, form.State, form.ZipCode)
	errors = append(errors, addressErrors...)

	// Validate payment method specific fields
	switch form.PaymentMethod {
	case "credit_card":
		cardErrors := ValidateCreditCard(form.CardName, form.CardNumber, form.Expiry, form.CVV)
		errors = append(errors, cardErrors...)
	case "boleto":
		if err := ValidateCPF(form.CPF); err != nil {
			errors = append(errors, *err)
		}
	case "pix":
		if err := ValidatePixKey(form.PixKey); err != nil {
			errors = append(errors, *err)
		}
	}

	// Validate cart items
	if len(form.CartItems) == 0 {
		errors = append(errors, ValidationError{Field: "cart", Message: "Seu carrinho está vazio"})
	}

	return ValidationResult{
		IsValid: len(errors) == 0,
		Errors:  errors,
	}
}

// GenerateOrderNumber generates a unique order number
func GenerateOrderNumber() string {
	timestamp := time.Now().UnixNano()
	random := fmt.Sprintf("%03d", timestamp%1000)
	return fmt.Sprintf("ORD-%d-%s", timestamp, random)
}

// CreateOrder creates a new order in the database
func CreateOrder(form CheckoutForm) (*Order, error) {
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	// Calculate total amount
	var totalAmount float64
	for _, item := range form.CartItems {
		totalAmount += item.Price * float64(item.Quantity)
	}

	orderNumber := GenerateOrderNumber()

	query := `
		INSERT INTO orders (
			order_number, email, phone, first_name, last_name, address, 
			neighborhood, city, state, zip_code, apartment, payment_method,
			total_amount, status
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		RETURNING id, created_at, updated_at
	`

	var order Order
	err := db.QueryRow(
		query,
		orderNumber,
		form.Email,
		form.Phone,
		form.FirstName,
		form.LastName,
		form.Address,
		form.Neighborhood,
		form.City,
		form.State,
		form.ZipCode,
		form.Apartment,
		form.PaymentMethod,
		totalAmount,
		"pending",
	).Scan(&order.ID, &order.CreatedAt, &order.UpdatedAt)

	if err != nil {
		return nil, fmt.Errorf("failed to create order: %v", err)
	}

	// Set order fields
	order.OrderNumber = orderNumber
	order.Email = form.Email
	order.Phone = form.Phone
	order.FirstName = form.FirstName
	order.LastName = form.LastName
	order.Address = form.Address
	order.Neighborhood = form.Neighborhood
	order.City = form.City
	order.State = form.State
	order.ZipCode = form.ZipCode
	order.Apartment = form.Apartment
	order.PaymentMethod = form.PaymentMethod
	order.TotalAmount = totalAmount
	order.Status = "pending"

	// Create order items
	for _, item := range form.CartItems {
		fmt.Printf("%v\n", item.ID)
		_, err := db.Exec(`
			INSERT INTO order_items (order_id, product_id, product_name, quantity, unit_price, total_price)
			VALUES ($1, $2, $3, $4, $5, $6)
		`, order.ID, item.ID, item.Name, item.Quantity, item.Price, item.Price*float64(item.Quantity))

		if err != nil {
			// Clean up order if item creation fails
			db.Exec("DELETE FROM orders WHERE id = $1", order.ID)
			return nil, fmt.Errorf("failed to create order item: %v", err)
		}
	}

	return &order, nil
}

// UpdateOrderPaymentStatus updates the payment status of an order
func UpdateOrderPaymentStatus(orderID int, paymentStatus, stripePaymentID string) error {
	if db == nil {
		return fmt.Errorf("database not initialized")
	}

	query := `
		UPDATE orders 
		SET payment_status = $1, stripe_payment_id = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $3
	`

	_, err := db.Exec(query, paymentStatus, stripePaymentID, orderID)
	return err
}

// GetOrderByID retrieves an order by ID
func GetOrderByID(orderID int) (*Order, error) {
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	query := `
		SELECT id, order_number, email, phone, first_name, last_name, address,
			   neighborhood, city, state, zip_code, apartment, payment_method,
			   payment_status, stripe_payment_id, total_amount, status, created_at, updated_at
		FROM orders WHERE id = $1
	`

	var order Order
	err := db.QueryRow(query, orderID).Scan(
		&order.ID, &order.OrderNumber, &order.Email, &order.Phone, &order.FirstName,
		&order.LastName, &order.Address, &order.Neighborhood, &order.City, &order.State,
		&order.ZipCode, &order.Apartment, &order.PaymentMethod, &order.PaymentStatus,
		&order.StripePaymentID, &order.TotalAmount, &order.Status, &order.CreatedAt, &order.UpdatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &order, nil
}

// GetOrderItems retrieves all items for an order
func GetOrderItems(orderID int) ([]OrderItem, error) {
	if db == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	query := `
		SELECT id, order_id, product_id, product_name, quantity, unit_price, total_price, created_at
		FROM order_items WHERE order_id = $1 ORDER BY id
	`

	rows, err := db.Query(query, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var items []OrderItem
	for rows.Next() {
		var item OrderItem
		err := rows.Scan(
			&item.ID, &item.OrderID, &item.ProductID, &item.ProductName,
			&item.Quantity, &item.UnitPrice, &item.TotalPrice, &item.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}

	return items, nil
}
