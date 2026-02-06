package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"lojagtec/internal/admin"
	"lojagtec/internal/checkout"
	"lojagtec/internal/database"
	"lojagtec/internal/orders"
	"lojagtec/internal/products"
)

const (
	maxUploadSize = 5 << 20 // 5MB
	uploadPath    = "web/static/images/uploads"
)

type adminDashboardData struct {
	CanViewOrders bool
	Brands        []products.Brand
	Products      []products.ProductOption
}

type adminEditData struct {
	Product         *products.Product
	Brands          []products.Brand
	Products        []products.ProductOption
	BrandSelections map[int]bool
	FitSelections   map[int]bool
}

type brandModalData struct {
	Name  string
	Error string
}

func main() {
	db, err := database.Connect()
	if err != nil {
		log.Fatalf("Could not connect to the database: %v", err)
	}
	defer db.Close()

	// Apply database schema
	if err := database.RunSchema(db); err != nil {
		log.Fatalf("Could not apply database schema: %v", err)
	}

	// Set database for packages
	products.SetDatabase(db)
	admin.SetDatabase(db)
	orders.SetDatabase(db)

	// Ensure upload directory exists
	if err := os.MkdirAll(uploadPath, 0755); err != nil {
		log.Fatalf("Could not create upload directory: %v", err)
	}

	fs := http.FileServer(http.Dir("web/static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	// Cart modal route (must be before catch-all "/")
	http.HandleFunc("/cart-modal", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("web/templates/cart-modal.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		tmpl.Execute(w, nil)
	})

	// Installation service modal route
	http.HandleFunc("/installation-service-modal", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("web/templates/installation-service-modal.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		tmpl.Execute(w, nil)
	})

	// Public routes
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("web/templates/index.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		tmpl.Execute(w, nil)
	})

	http.HandleFunc("/checkout", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("web/templates/checkout.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		tmpl.Execute(w, nil)
	})

	http.HandleFunc("/checkout/success", func(w http.ResponseWriter, r *http.Request) {
		orderID, err := strconv.Atoi(r.URL.Query().Get("order_id"))
		if err != nil {
			http.Error(w, "Invalid order ID", http.StatusBadRequest)
			return
		}

		order, err := orders.GetOrderByID(orderID)
		if err != nil {
			http.Error(w, "Pedido não encontrado", http.StatusNotFound)
			return
		}

		tmpl, err := template.ParseFiles("web/templates/checkout-success-page.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		tmpl.Execute(w, map[string]interface{}{
			"Order": order,
		})
	})

	http.HandleFunc("/checkout/cancel", func(w http.ResponseWriter, r *http.Request) {
		orderID, err := strconv.Atoi(r.URL.Query().Get("order_id"))
		if err != nil {
			http.Error(w, "Invalid order ID", http.StatusBadRequest)
			return
		}

		order, err := orders.GetOrderByID(orderID)
		if err != nil {
			http.Error(w, "Pedido não encontrado", http.StatusNotFound)
			return
		}

		tmpl, err := template.ParseFiles("web/templates/checkout-cancel-page.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		tmpl.Execute(w, map[string]interface{}{
			"Order": order,
		})
	})

	// Checkout endpoint for order processing
	http.HandleFunc("/api/checkout", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Parse form data
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}

		// Create checkout form from request data
		form := orders.CheckoutForm{
			Email:         r.FormValue("email"),
			Phone:         r.FormValue("phone"),
			FirstName:     r.FormValue("firstName"),
			LastName:      r.FormValue("lastName"),
			Address:       r.FormValue("address"),
			Neighborhood:  r.FormValue("neighborhood"),
			City:          r.FormValue("city"),
			State:         r.FormValue("state"),
			ZipCode:       r.FormValue("zipCode"),
			Apartment:     r.FormValue("apartment"),
			CPF:           r.FormValue("cpf"),
			PaymentMethod: r.FormValue("paymentMethod"),
		}

		// Parse cart items from form data
		cartData := r.FormValue("cart_items")
		if cartData == "" {
			// Return error for empty cart
			tmpl, err := template.ParseFiles("web/templates/validation-error.html")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			tmpl.Execute(w, orders.ValidationError{Field: "cart", Message: "Seu carrinho está vazio"})
			return
		}

		if err := json.Unmarshal([]byte(cartData), &form.CartItems); err != nil {
			http.Error(w, "Invalid cart data", http.StatusBadRequest)
			return
		}

		// Validate the form
		result := orders.ValidateCheckoutForm(form)
		if !result.IsValid {
			// Return validation errors
			tmpl, err := template.ParseFiles("web/templates/validation-error.html")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			tmpl.Execute(w, result.Errors[0])
			return
		}

		// Create the order
		order, err := orders.CreateOrder(form)
		if err != nil {
			tmpl, parseErr := template.ParseFiles("web/templates/validation-error.html")
			if parseErr != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			if errors.Is(err, orders.ErrInstallationServiceUnavailable) || errors.Is(err, orders.ErrInvalidCartItem) {
				tmpl.Execute(w, orders.ValidationError{Field: "cart", Message: err.Error()})
				return
			}
			tmpl.Execute(w, orders.ValidationError{Field: "general", Message: "Erro ao processar pedido: " + err.Error()})
			return
		}

		stripeSessionURL, err := checkout.CreateCheckoutSession(form, order)
		if err != nil {
			var validationErr checkout.ValidationError
			if errors.Is(err, checkout.ErrStripeNotConfigured) {
				tmpl, parseErr := template.ParseFiles("web/templates/validation-error.html")
				if parseErr != nil {
					http.Error(w, "Stripe configuration error", http.StatusInternalServerError)
					return
				}
				tmpl.Execute(w, orders.ValidationError{Field: "general", Message: "Pagamento temporariamente indisponível."})
				return
			}
			if errors.As(err, &validationErr) {
				tmpl, parseErr := template.ParseFiles("web/templates/validation-error.html")
				if parseErr != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				tmpl.Execute(w, orders.ValidationError{Field: validationErr.Field, Message: validationErr.Message})
				return
			}

			log.Printf("Failed to create Stripe session: %v", err)
			tmpl, parseErr := template.ParseFiles("web/templates/validation-error.html")
			if parseErr != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			tmpl.Execute(w, orders.ValidationError{Field: "general", Message: "Erro ao iniciar pagamento. Tente novamente."})
			return
		}

		w.Header().Set("HX-Redirect", stripeSessionURL)
		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("/api/stripe/webhook", func(w http.ResponseWriter, r *http.Request) {
		checkout.HandleStripeWebhook(w, r)
	})

	// Product filter routes
	http.HandleFunc("/products/bebedouros", func(w http.ResponseWriter, r *http.Request) {
		prods, err := products.GetProductsByCategory("bebedouros")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		templateFile := "web/templates/product-cards.html"
		if len(prods) == 0 {
			templateFile = "web/templates/product-empty-state.html"
		}

		tmpl, err := template.ParseFiles(templateFile)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		tmpl.Execute(w, prods)
	})

	http.HandleFunc("/products/purificadores", func(w http.ResponseWriter, r *http.Request) {
		prods, err := products.GetProductsByCategory("purificadores")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		templateFile := "web/templates/product-cards.html"
		if len(prods) == 0 {
			templateFile = "web/templates/product-empty-state.html"
		}

		tmpl, err := template.ParseFiles(templateFile)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, prods)
	})

	http.HandleFunc("/products/refis", func(w http.ResponseWriter, r *http.Request) {
		prods, err := products.GetProductsByCategory("refis")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		templateFile := "web/templates/product-cards.html"
		if len(prods) == 0 {
			templateFile = "web/templates/product-empty-state.html"
		}

		tmpl, err := template.ParseFiles(templateFile)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, prods)
	})

	http.HandleFunc("/products/pecas", func(w http.ResponseWriter, r *http.Request) {
		prods, err := products.GetProductsByCategory("pecas")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		templateFile := "web/templates/product-cards.html"
		if len(prods) == 0 {
			templateFile = "web/templates/product-empty-state.html"
		}

		tmpl, err := template.ParseFiles(templateFile)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, prods)
	})

	http.HandleFunc("/products/all", func(w http.ResponseWriter, r *http.Request) {
		prods, err := products.GetAllProducts()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		templateFile := "web/templates/product-cards.html"
		if len(prods) == 0 {
			templateFile = "web/templates/product-empty-state.html"
		}

		tmpl, err := template.ParseFiles(templateFile)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, prods)
	})

	// Admin routes
	http.HandleFunc("/admin/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			// Check if already authenticated
			if admin.IsAuthenticated(r) {
				http.Redirect(w, r, "/admin", http.StatusSeeOther)
				return
			}

			tmpl, err := template.ParseFiles("web/templates/admin-login.html")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			tmpl.Execute(w, nil)
			return
		}

		if r.Method == http.MethodPost {
			username := r.FormValue("username")
			password := r.FormValue("password")

			err := admin.Login(w, username, password)
			if err != nil {
				tmpl, _ := template.ParseFiles("web/templates/admin-login.html")
				tmpl.Execute(w, map[string]string{"Error": err.Error()})
				return
			}

			http.Redirect(w, r, "/admin", http.StatusSeeOther)
			return
		}

		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	})

	http.HandleFunc("/admin/logout", func(w http.ResponseWriter, r *http.Request) {
		admin.Logout(w, r)
		http.Redirect(w, r, "/admin/login", http.StatusSeeOther)
	})

	http.HandleFunc("/admin", admin.RequireRole("admin", "product_admin")(func(w http.ResponseWriter, r *http.Request) {
		brands, err := products.GetAllBrands()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		productOptions, err := products.GetAllProductOptions()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tmpl, err := template.ParseFiles("web/templates/admin-dashboard.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, adminDashboardData{
			CanViewOrders: true,
			Brands:        brands,
			Products:      productOptions,
		})
	}))

	http.HandleFunc("/admin/brands/new", admin.RequireRole("admin", "product_admin")(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		tmpl, err := template.ParseFiles("web/templates/admin-brand-modal.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, brandModalData{})
	}))

	http.HandleFunc("/admin/orders", admin.RequireRole("admin", "product_admin")(func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("web/templates/admin-orders.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, nil)
	}))

	// Admin API routes
	http.HandleFunc("/api/admin/orders", admin.RequireRole("admin", "product_admin")(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		status := r.URL.Query().Get("status")
		paymentStatus := r.URL.Query().Get("payment_status")

		filters := orders.OrderFilters{
			Status:        status,
			PaymentStatus: paymentStatus,
			Limit:         50,
			Offset:        0,
		}

		ordersList, err := orders.GetOrders(filters)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		role, _ := admin.RoleFromRequest(r)
		canViewFinancialData := role == "admin"

		w.Header().Set("Content-Type", "text/html")
		tmpl, err := template.ParseFiles("web/templates/admin-orders-list.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		data := map[string]interface{}{
			"Orders":               ordersList,
			"CanViewFinancialData": canViewFinancialData,
		}

		if canViewFinancialData {
			totals, err := orders.GetOrderTotals(filters)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			data["Totals"] = totals
		}

		tmpl.Execute(w, data)
	}))

	http.HandleFunc("/api/admin/brands/options", admin.RequireRole("admin", "product_admin")(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		brands, err := products.GetAllBrands()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tmpl, err := template.ParseFiles("web/templates/admin-brand-options.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html")
		tmpl.Execute(w, brands)
	}))

	http.HandleFunc("/api/admin/brands", admin.RequireRole("admin", "product_admin")(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Invalid form data", http.StatusBadRequest)
			return
		}
		name := r.FormValue("name")
		brand, err := products.CreateBrand(name)
		if err != nil {
			if r.Header.Get("HX-Request") == "true" {
				tmpl, tmplErr := template.ParseFiles("web/templates/admin-brand-modal.html")
				if tmplErr != nil {
					http.Error(w, tmplErr.Error(), http.StatusInternalServerError)
					return
				}
				w.WriteHeader(http.StatusBadRequest)
				tmpl.Execute(w, brandModalData{Name: name, Error: err.Error()})
				return
			}
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		if r.Header.Get("HX-Request") == "true" {
			w.Header().Set("HX-Trigger", "refreshBrands,closeBrandModal")
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(brand)
	}))

	http.HandleFunc("/api/admin/orders/", admin.RequireRole("admin", "product_admin")(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		path := strings.TrimPrefix(r.URL.Path, "/api/admin/orders/")
		id, err := strconv.Atoi(path)
		if err != nil {
			http.Error(w, "Invalid order ID", http.StatusBadRequest)
			return
		}

		order, items, err := orders.GetOrderWithItems(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		role, _ := admin.RoleFromRequest(r)
		canViewFinancialData := role == "admin"

		w.Header().Set("Content-Type", "text/html")
		tmpl, err := template.ParseFiles("web/templates/admin-order-detail.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, map[string]interface{}{
			"Order":                order,
			"Items":                items,
			"CanViewFinancialData": canViewFinancialData,
		})
	}))

	http.HandleFunc("/api/admin/products", admin.RequireRole("admin", "product_admin")(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			// Check if this is an HTMX request for HTML fragment
			if r.Header.Get("HX-Request") == "true" {
				// Create a simple function map for template
				funcMap := template.FuncMap{
					"formatCategory": func(category string) string {
						categories := map[string]string{
							"bebedouros":    "Bebedouros",
							"purificadores": "Purificadores",
							"refis":         "Refis",
							"pecas":         "Peças",
						}
						if val, ok := categories[category]; ok {
							return val
						}
						return category
					},
				}

				tmpl, err := template.New("admin-product-list.html").Funcs(funcMap).ParseFiles("web/templates/admin-product-list.html")
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				prods, err := products.GetAllProducts()
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				w.Header().Set("Content-Type", "text/html")
				tmpl.Execute(w, prods)
			} else {
				// Original JSON response for non-HTMX requests
				w.Header().Set("Content-Type", "application/json")
				prods, err := products.GetAllProducts()
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				json.NewEncoder(w).Encode(prods)
			}

		case http.MethodPost:
			// Parse multipart form (max 5MB)
			if err := r.ParseMultipartForm(maxUploadSize); err != nil {
				http.Error(w, "File too large or invalid form data", http.StatusBadRequest)
				return
			}

			// Get form values
			name := r.FormValue("name")
			priceStr := r.FormValue("price")
			category := r.FormValue("category")
			isAvailableStr := r.FormValue("is_available")
			brandIDs, err := parseIDList(r.Form["brand_ids"])
			if err != nil {
				http.Error(w, "Invalid brand selection", http.StatusBadRequest)
				return
			}
			fitsProductIDs, err := parseIDList(r.Form["fits_product_ids"])
			if err != nil {
				http.Error(w, "Invalid compatibility selection", http.StatusBadRequest)
				return
			}

			if name == "" || priceStr == "" || category == "" {
				http.Error(w, "Missing required fields", http.StatusBadRequest)
				return
			}

			if len(fitsProductIDs) > 0 && !isPartsCategory(category) {
				http.Error(w, "Compatibility is only allowed for refis or pecas", http.StatusBadRequest)
				return
			}

			price, err := strconv.ParseFloat(priceStr, 64)
			if err != nil {
				http.Error(w, "Invalid price format", http.StatusBadRequest)
				return
			}

			// Handle file upload
			imagePath, err := handleImageUpload(r, "image")
			if err != nil {
				http.Error(w, fmt.Sprintf("Image upload failed: %v", err), http.StatusBadRequest)
				return
			}

			isAvailable := isAvailableStr == "on"
			// Create product
			product, err := products.CreateProduct(name, price, imagePath, category, isAvailable, brandIDs, fitsProductIDs)
			if err != nil {
				// Clean up uploaded file if product creation fails
				os.Remove(filepath.Join(uploadPath, filepath.Base(imagePath)))

				// Return error message for HTMX
				if r.Header.Get("HX-Request") == "true" {
					tmpl, _ := template.ParseFiles("web/templates/admin-error-message.html")
					tmpl.Execute(w, err.Error())
				} else {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
				return
			}

			// Return product card HTML for HTMX
			if r.Header.Get("HX-Request") == "true" {
				// Create function map for template
				funcMap := template.FuncMap{
					"formatCategory": func(category string) string {
						categories := map[string]string{
							"bebedouros":    "Bebedouros",
							"purificadores": "Purificadores",
							"refis":         "Refis",
							"pecas":         "Peças",
						}
						if val, ok := categories[category]; ok {
							return val
						}
						return category
					},
				}

				tmpl, err := template.New("admin-product-card.html").Funcs(funcMap).ParseFiles("web/templates/admin-product-card.html")
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				w.Header().Set("Content-Type", "text/html")
				tmpl.Execute(w, product)
			} else {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusCreated)
				json.NewEncoder(w).Encode(product)
			}

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	http.HandleFunc("/api/admin/products/", admin.RequireRole("admin", "product_admin")(func(w http.ResponseWriter, r *http.Request) {
		// Extract ID from path
		path := strings.TrimPrefix(r.URL.Path, "/api/admin/products/")
		id, err := strconv.Atoi(path)
		if err != nil {
			http.Error(w, "Invalid product ID", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodPut:
			// Parse multipart form
			if err := r.ParseMultipartForm(maxUploadSize); err != nil {
				http.Error(w, "File too large or invalid form data", http.StatusBadRequest)
				return
			}

			// Get form values
			name := r.FormValue("name")
			priceStr := r.FormValue("price")
			category := r.FormValue("category")
			currentImage := r.FormValue("current_image")
			isAvailableStr := r.FormValue("is_available")
			brandIDs, err := parseIDList(r.Form["brand_ids"])
			if err != nil {
				http.Error(w, "Invalid brand selection", http.StatusBadRequest)
				return
			}
			fitsProductIDs, err := parseIDList(r.Form["fits_product_ids"])
			if err != nil {
				http.Error(w, "Invalid compatibility selection", http.StatusBadRequest)
				return
			}

			if name == "" || priceStr == "" || category == "" {
				http.Error(w, "Missing required fields", http.StatusBadRequest)
				return
			}

			if len(fitsProductIDs) > 0 && !isPartsCategory(category) {
				http.Error(w, "Compatibility is only allowed for refis or pecas", http.StatusBadRequest)
				return
			}

			price, err := strconv.ParseFloat(priceStr, 64)
			if err != nil {
				http.Error(w, "Invalid price format", http.StatusBadRequest)
				return
			}

			// Parse is_available checkbox (unchecked checkboxes are not sent in form data)
			isAvailable := isAvailableStr == "on"

			// Check if new image was uploaded
			imagePath := currentImage
			file, _, err := r.FormFile("image")
			if err == nil {
				// New image uploaded
				file.Close()
				newImagePath, err := handleImageUpload(r, "image")
				if err != nil {
					http.Error(w, fmt.Sprintf("Image upload failed: %v", err), http.StatusBadRequest)
					return
				}
				imagePath = newImagePath

				// Delete old image if it's in uploads folder
				if strings.Contains(currentImage, "/uploads/") {
					oldImagePath := filepath.Join("web/static", strings.TrimPrefix(currentImage, "/static/"))
					os.Remove(oldImagePath)
				}
			}

			// Update product
			if err := products.UpdateProduct(id, name, price, imagePath, category, isAvailable, brandIDs, fitsProductIDs); err != nil {
				// Return error message for HTMX
				if r.Header.Get("HX-Request") == "true" {
					tmpl, _ := template.ParseFiles("web/templates/admin-error-message.html")
					tmpl.Execute(w, err.Error())
				} else {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
				return
			}

			// Return updated product card HTML for HTMX
			if r.Header.Get("HX-Request") == "true" {
				// Get the updated product data
				updatedProduct, err := products.GetProductByID(id)
				if err != nil {
					tmpl, _ := template.ParseFiles("web/templates/admin-error-message.html")
					tmpl.Execute(w, err.Error())
					return
				}

				// Create function map for template
				funcMap := template.FuncMap{
					"formatCategory": func(category string) string {
						categories := map[string]string{
							"bebedouros":    "Bebedouros",
							"purificadores": "Purificadores",
							"refis":         "Refis",
							"pecas":         "Peças",
						}
						if val, ok := categories[category]; ok {
							return val
						}
						return category
					},
				}

				tmpl, err := template.New("admin-product-card.html").Funcs(funcMap).ParseFiles("web/templates/admin-product-card.html")
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				w.Header().Set("Content-Type", "text/html")
				w.Header().Set("HX-Trigger", "closeEditModal")
				tmpl.Execute(w, updatedProduct)
			} else {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]string{"message": "Product updated successfully"})
			}

		case http.MethodDelete:
			// Get product to find image path
			product, err := products.GetProductByID(id)
			if err == nil && strings.Contains(product.Image, "/uploads/") {
				// Delete image file if it's in uploads folder
				imagePath := filepath.Join("web/static", strings.TrimPrefix(product.Image, "/static/"))
				os.Remove(imagePath)
			}

			if err := products.DeleteProduct(id); err != nil {
				// Return error message for HTMX
				if r.Header.Get("HX-Request") == "true" {
					tmpl, _ := template.ParseFiles("web/templates/admin-error-message.html")
					tmpl.Execute(w, err.Error())
				} else {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
				return
			}

			// Return empty response for HTMX (element will be deleted)
			if r.Header.Get("HX-Request") == "true" {
				w.WriteHeader(http.StatusOK)
			} else {
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]string{"message": "Product deleted successfully"})
			}

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	// HTMX-specific admin routes
	http.HandleFunc("/admin/products/", admin.RequireRole("admin", "product_admin")(func(w http.ResponseWriter, r *http.Request) {
		// Extract ID from path
		path := strings.TrimPrefix(r.URL.Path, "/admin/products/")
		parts := strings.Split(path, "/")
		if len(parts) < 1 {
			http.Error(w, "Invalid path", http.StatusBadRequest)
			return
		}

		id, err := strconv.Atoi(parts[0])
		if err != nil {
			http.Error(w, "Invalid product ID", http.StatusBadRequest)
			return
		}

		// Handle edit form request
		if len(parts) >= 2 && parts[1] == "edit" {
			product, err := products.GetProductByID(id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusNotFound)
				return
			}
			brands, err := products.GetAllBrands()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			productOptions, err := products.GetAllProductOptions()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			filteredOptions := make([]products.ProductOption, 0, len(productOptions))
			for _, option := range productOptions {
				if option.ID == product.ProductID {
					continue
				}
				filteredOptions = append(filteredOptions, option)
			}

			tmpl, err := template.ParseFiles("web/templates/admin-edit-form.html")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			editData := adminEditData{
				Product:         product,
				Brands:          brands,
				Products:        filteredOptions,
				BrandSelections: buildIDSet(product.BrandIDs),
				FitSelections:   buildIDSet(product.FitsProductIDs),
			}
			tmpl.Execute(w, editData)
			return
		}

		http.Error(w, "Not found", http.StatusNotFound)
	}))

	fmt.Println("Server starting at port 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("Error starting server: %s\n", err)
	}
}

// handleImageUpload processes the uploaded image file
func handleImageUpload(r *http.Request, fieldName string) (string, error) {
	file, header, err := r.FormFile(fieldName)
	if err != nil {
		return "", fmt.Errorf("failed to get file from form: %v", err)
	}
	defer file.Close()

	// Validate file size
	if header.Size > maxUploadSize {
		return "", fmt.Errorf("file size exceeds maximum allowed size of 5MB")
	}

	// Validate file type
	contentType := header.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		return "", fmt.Errorf("file must be an image")
	}

	// Generate unique filename
	ext := filepath.Ext(header.Filename)
	randomBytes := make([]byte, 16)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("failed to generate random filename: %v", err)
	}
	filename := hex.EncodeToString(randomBytes) + ext

	// Create file path
	filePath := filepath.Join(uploadPath, filename)

	// Create destination file
	dst, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %v", err)
	}
	defer dst.Close()

	// Copy uploaded file to destination
	if _, err := io.Copy(dst, file); err != nil {
		os.Remove(filePath)
		return "", fmt.Errorf("failed to save file: %v", err)
	}

	// Return the web-accessible path
	return "/static/images/uploads/" + filename, nil
}

func parseIDList(values []string) ([]int, error) {
	ids := make([]int, 0, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		id, err := strconv.Atoi(value)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func buildIDSet(ids []int) map[int]bool {
	set := make(map[int]bool, len(ids))
	for _, id := range ids {
		set[id] = true
	}
	return set
}

func isPartsCategory(category string) bool {
	return category == "refis" || category == "pecas"
}
