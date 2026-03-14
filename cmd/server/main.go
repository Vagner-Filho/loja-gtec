package main

import (
	"crypto/rand"
	"database/sql"
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
	"time"

	"lojagtec/internal/admin"
	"lojagtec/internal/banners"
	"lojagtec/internal/checkout"
	"lojagtec/internal/database"
	"lojagtec/internal/offers"
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
	ProductImages   []products.ProductImage
}

type brandModalData struct {
	Name  string
	Error string
}

type productPageData struct {
	Product            *products.Product
	Brands             []products.Brand
	CompatibleProducts []products.Product
	RelatedProducts    []products.Product
}

// setCacheHeaders sets HTTP cache headers for HTMX modal responses
func setCacheHeaders(w http.ResponseWriter, maxAgeSeconds int) {
	w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", maxAgeSeconds))
	w.Header().Set("Expires", time.Now().Add(time.Duration(maxAgeSeconds)*time.Second).Format(http.TimeFormat))
}

// translateStatus translates order status from English to Portuguese
func translateStatus(status string) string {
	switch status {
	case "pending":
		return "Pendente"
	case "processing":
		return "Em processamento"
	case "shipped":
		return "Enviado"
	case "completed":
		return "Concluído"
	case "cancelled":
		return "Cancelado"
	default:
		return status
	}
}

// translatePaymentStatus translates payment status from English to Portuguese
func translatePaymentStatus(status string) string {
	switch status {
	case "paid":
		return "Pago"
	case "pending":
		return "Pendente"
	case "failed":
		return "Falhou"
	case "waiting":
		return "Aguardando"
	default:
		return status
	}
}

// translatePaymentMethod translates payment method from English to Portuguese
func translatePaymentMethod(method string) string {
	switch method {
	case "credit_card":
		return "Cartão de Crédito"
	case "boleto":
		return "Boleto"
	case "pix":
		return "PIX"
	default:
		return method
	}
}

// parseBrandIDs parses a comma-separated string of brand IDs into a slice of ints
func parseBrandIDs(brands string) []int {
	if brands == "" {
		return nil
	}

	parts := strings.Split(brands, ",")
	var brandIDs []int
	for _, part := range parts {
		id, err := strconv.Atoi(strings.TrimSpace(part))
		if err == nil && id > 0 {
			brandIDs = append(brandIDs, id)
		}
	}
	return brandIDs
}

// orderFuncMap returns a template.FuncMap with order-related helper functions
func orderFuncMap() template.FuncMap {
	return template.FuncMap{
		"translateStatus":        translateStatus,
		"translatePaymentStatus": translatePaymentStatus,
		"translatePaymentMethod": translatePaymentMethod,
		"sub": func(a, b float64) float64 {
			return a - b
		},
	}
}

// cacheControlWrapper wraps a handler to add cache control headers
type cacheControlWrapper struct {
	handler http.Handler
	maxAge  int
}

func (w cacheControlWrapper) ServeHTTP(res http.ResponseWriter, req *http.Request) {
	setCacheHeaders(res, w.maxAge)
	w.handler.ServeHTTP(res, req)
}

func main() {
	db, err := database.Connect()
	if err != nil {
		log.Fatalf("Could not connect to the database: %v", err)
	}
	defer db.Close()

	// Set database for packages
	products.SetDatabase(db)
	admin.SetDatabase(db)
	orders.SetDatabase(db)
	banners.SetDatabase(db)
	offers.SetDatabase(db)

	// Apply database schema
	if err := database.RunSchema(db); err != nil {
		log.Fatalf("Could not apply database schema: %v", err)
	}

	// Ensure upload directory exists
	if err := os.MkdirAll(uploadPath, 0755); err != nil {
		log.Fatalf("Could not create upload directory: %v", err)
	}

	fs := http.FileServer(http.Dir("web/static"))
	http.Handle("/static/", http.StripPrefix("/static/", cacheControlWrapper{handler: fs, maxAge: 86400}))

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
		brandIDs := parseBrandIDs(r.URL.Query().Get("brands"))
		prods, err := products.GetProductsByCategoryAndBrands("bebedouros", brandIDs)
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
		brandIDs := parseBrandIDs(r.URL.Query().Get("brands"))
		prods, err := products.GetProductsByCategoryAndBrands("purificadores", brandIDs)
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
		brandIDs := parseBrandIDs(r.URL.Query().Get("brands"))
		prods, err := products.GetProductsByCategoryAndBrands("refis", brandIDs)
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
		brandIDs := parseBrandIDs(r.URL.Query().Get("brands"))
		prods, err := products.GetProductsByCategoryAndBrands("pecas", brandIDs)
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
		brandIDs := parseBrandIDs(r.URL.Query().Get("brands"))
		prods, err := products.GetProductsByCategoryAndBrands("", brandIDs)
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
		err = tmpl.Execute(w, prods)
	})

	// Offers route - displays products currently on offer
	http.HandleFunc("/products/ofertas", func(w http.ResponseWriter, r *http.Request) {
		offerProducts, err := offers.GetActiveOffers()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Convert offers to products format for template
		prods := make([]products.Product, len(offerProducts))
		for i, offer := range offerProducts {
			// Get primary image from product_images
			image, _ := products.GetPrimaryProductImage(offer.ProductID)
			prods[i] = products.Product{
				ID:             offer.ID,
				ProductID:      offer.ProductID,
				Name:           offer.Name,
				Price:          offer.Price,
				Image:          image,
				Category:       offer.Category,
				IsOnOffer:      true,
				OfferPrice:     offer.OfferPrice,
				OfferStartDate: offer.StartDate,
				OfferEndDate:   offer.EndDate,
				IsAvailable:    true,
			}
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

	// Brands endpoint for filter UI
	http.HandleFunc("/api/brands", func(w http.ResponseWriter, r *http.Request) {
		brands, err := products.GetAllBrands()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html")
		for _, brand := range brands {
			fmt.Fprintf(w, `<button type="button" 
				data-brand-id="%d" 
				class="brand-pill px-3 py-1 rounded-full text-sm transition-colors duration-200 bg-gray-200 text-gray-700 hover:bg-gray-300"
				onclick="toggleBrandFilter(this)">
				%s
			</button>`, brand.ID, brand.Name)
		}
	})

	// Search endpoint for fuzzy product and brand search
	http.HandleFunc("/api/search", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("q")
		if query == "" {
			// Return empty results if no query
			tmpl, err := template.ParseFiles("web/templates/search-results.html")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			tmpl.Execute(w, map[string]interface{}{
				"Query":    "",
				"Products": []products.Product{},
				"Brands":   []products.BrandSearchResult{},
			})
			return
		}

		// Search for products (max 6 results)
		productResults, err := products.SearchProducts(query, 6)
		if err != nil {
			log.Printf("Error searching products: %v", err)
			http.Error(w, "Error searching products", http.StatusInternalServerError)
			return
		}

		// Search for brands (max 6 results)
		brandResults, err := products.SearchBrandsWithCount(query, 6)
		if err != nil {
			log.Printf("Error searching brands: %v", err)
			http.Error(w, "Error searching brands", http.StatusInternalServerError)
			return
		}

		// Prepare data for template
		searchData := struct {
			Query    string
			Products []products.Product
			Brands   []products.BrandSearchResult
		}{
			Query:    query,
			Products: productResults,
			Brands:   brandResults,
		}

		// Render the search results template
		tmpl, err := template.ParseFiles("web/templates/search-results.html")
		if err != nil {
			log.Printf("Error parsing search results template: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html")
		tmpl.Execute(w, searchData)
	})

	// Product detail page
	http.HandleFunc("/produto/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Extract product ID from URL
		path := strings.TrimPrefix(r.URL.Path, "/produto/")
		parts := strings.Split(path, "/")
		if len(parts) == 0 || parts[0] == "" {
			http.Error(w, "Invalid product ID", http.StatusBadRequest)
			return
		}

		productID, err := strconv.Atoi(parts[0])
		if err != nil {
			http.Error(w, "Invalid product ID", http.StatusBadRequest)
			return
		}

		// Get product from database
		product, err := products.GetProductByID(productID)
		if err != nil {
			// Check if product not found
			if errors.Is(err, sql.ErrNoRows) {
				http.Error(w, "Product not found", http.StatusNotFound)
				return
			}
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Rule 1: Block unavailable products
		if !product.IsAvailable {
			// Return 404 page
			tmpl, err := template.ParseFiles("web/templates/404.html")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.WriteHeader(http.StatusNotFound)
			tmpl.Execute(w, nil)
			return
		}

		// Get brand names for this product
		brandNames, err := products.GetBrandNamesByProductID(product.ProductID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Get compatible products (for refis and pecas)
		var compatibleProducts []products.Product
		if product.Category == "refis" || product.Category == "pecas" {
			compatibleProducts, err = products.GetCompatibleProductsByProductID(product.ProductID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
		}

		// Get related products (same category, excluding current product)
		relatedProducts, err := products.GetRelatedProducts(product.ProductID, product.Category, 8)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Parse template with custom function map
		tmpl, err := template.New("product.html").Funcs(template.FuncMap{
			"sub": func(a, b float64) float64 {
				return a - b
			},
		}).ParseFiles("web/templates/product.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		data := productPageData{
			Product:            product,
			Brands:             brandNames,
			CompatibleProducts: compatibleProducts,
			RelatedProducts:    relatedProducts,
		}

		tmpl.Execute(w, data)
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

	http.HandleFunc("/admin/banners", admin.RequireRole("admin", "product_admin")(func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("web/templates/admin-banners.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, nil)
	}))

	http.HandleFunc("/admin/offers", admin.RequireRole("admin", "product_admin")(func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("web/templates/admin-offers.html")
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
		funcMap := orderFuncMap()
		tmpl, err := template.New("admin-orders-list.html").Funcs(funcMap).ParseFiles("web/templates/admin-orders-list.html")
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
				setCacheHeaders(w, 86400) // 1 day cache for modal template
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

	// Banner management routes
	http.HandleFunc("/api/admin/banners", admin.RequireRole("admin", "product_admin")(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			// Return HTML list for HTMX requests
			if r.Header.Get("HX-Request") == "true" {
				tmpl, err := template.ParseFiles("web/templates/admin-banner-list.html")
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				bannerList, err := banners.GetAllBanners()
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				w.Header().Set("Content-Type", "text/html")
				tmpl.Execute(w, bannerList)
			} else {
				// Return JSON for non-HTMX requests
				w.Header().Set("Content-Type", "application/json")
				bannerList, err := banners.GetAllBanners()
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				json.NewEncoder(w).Encode(bannerList)
			}

		case http.MethodPost:
			// Parse multipart form (max 5MB)
			if err := r.ParseMultipartForm(maxUploadSize); err != nil {
				http.Error(w, "File too large or invalid form data", http.StatusBadRequest)
				return
			}

			// Get form values
			title := r.FormValue("title")
			linkURL := r.FormValue("link_url")

			// Validate title is required
			if title == "" {
				http.Error(w, "Title is required", http.StatusBadRequest)
				return
			}

			// Handle file upload
			imagePath, err := handleImageUpload(r, "image")
			if err != nil {
				http.Error(w, fmt.Sprintf("Image upload failed: %v", err), http.StatusBadRequest)
				return
			}

			// Create banner
			banner, err := banners.CreateBanner(imagePath, title, linkURL)
			if err != nil {
				// Clean up uploaded file on failure
				os.Remove(filepath.Join(uploadPath, filepath.Base(imagePath)))
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if r.Header.Get("HX-Request") == "true" {
				w.Header().Set("HX-Trigger", "refreshBanners")
				w.WriteHeader(http.StatusCreated)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(banner)

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	// Delete banner route
	http.HandleFunc("/api/admin/banners/", admin.RequireRole("admin", "product_admin")(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/admin/banners/")

		// Handle toggle status: /api/admin/banners/{id}/toggle
		if strings.HasSuffix(path, "/toggle") {
			if r.Method != http.MethodPut {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}

			idStr := strings.TrimSuffix(path, "/toggle")
			id, err := strconv.Atoi(idStr)
			if err != nil {
				http.Error(w, "Invalid banner ID", http.StatusBadRequest)
				return
			}

			isActive, err := banners.ToggleBannerStatus(id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if r.Header.Get("HX-Request") == "true" {
				w.Header().Set("HX-Trigger", "refreshBanners")
				w.WriteHeader(http.StatusOK)
				return
			}
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]bool{"is_active": isActive})
			return
		}

		// Handle reorder: /api/admin/banners/{id}/reorder
		if strings.HasSuffix(path, "/reorder") {
			if r.Method != http.MethodPost {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
				return
			}

			idStr := strings.TrimSuffix(path, "/reorder")
			id, err := strconv.Atoi(idStr)
			if err != nil {
				http.Error(w, "Invalid banner ID", http.StatusBadRequest)
				return
			}

			if err := r.ParseForm(); err != nil {
				http.Error(w, "Invalid form data", http.StatusBadRequest)
				return
			}

			newOrder, err := strconv.Atoi(r.FormValue("order"))
			if err != nil {
				http.Error(w, "Invalid order value", http.StatusBadRequest)
				return
			}

			if err := banners.UpdateBannerOrder(id, newOrder); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			if r.Header.Get("HX-Request") == "true" {
				w.Header().Set("HX-Trigger", "refreshBanners")
				w.WriteHeader(http.StatusOK)
				return
			}
			w.WriteHeader(http.StatusOK)
			return
		}

		// Handle delete: /api/admin/banners/{id}
		if r.Method != http.MethodDelete {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		id, err := strconv.Atoi(path)
		if err != nil {
			http.Error(w, "Invalid banner ID", http.StatusBadRequest)
			return
		}

		// Get banner to delete its image file
		banner, err := banners.GetBannerByID(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		// Delete from database
		if err := banners.DeleteBanner(id); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Delete image file
		os.Remove(filepath.Join(uploadPath, filepath.Base(banner.ImagePath)))

		if r.Header.Get("HX-Request") == "true" {
			w.Header().Set("HX-Trigger", "refreshBanners")
			w.WriteHeader(http.StatusOK)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))

	// Public banner endpoint - returns carousel HTML
	http.HandleFunc("/api/banners", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		tmpl, err := template.ParseFiles("web/templates/carousel.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		bannerList, err := banners.GetActiveBanners()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html")
		tmpl.Execute(w, bannerList)
	})

	// Admin offers management routes
	http.HandleFunc("/api/admin/offers", admin.RequireRole("admin", "product_admin")(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			// Return HTML list for HTMX requests
			if r.Header.Get("HX-Request") == "true" {
				tmpl, err := template.ParseFiles("web/templates/admin-offers-list.html")
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				offerList, err := offers.GetAllOffers()
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				w.Header().Set("Content-Type", "text/html")
				tmpl.Execute(w, offerList)
			} else {
				// Return JSON for non-HTMX requests
				w.Header().Set("Content-Type", "application/json")
				offerList, err := offers.GetAllOffers()
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				json.NewEncoder(w).Encode(offerList)
			}

		case http.MethodPost:
			if err := r.ParseForm(); err != nil {
				http.Error(w, "Invalid form data", http.StatusBadRequest)
				return
			}

			productID, err := strconv.Atoi(r.FormValue("product_id"))
			if err != nil {
				http.Error(w, "Invalid product ID", http.StatusBadRequest)
				return
			}

			offerPrice, err := strconv.ParseFloat(r.FormValue("offer_price"), 64)
			if err != nil {
				http.Error(w, "Invalid offer price", http.StatusBadRequest)
				return
			}

			form := offers.OfferForm{
				ProductID:  productID,
				OfferPrice: offerPrice,
			}

			// Parse optional dates
			startDateStr := r.FormValue("offer_start_date")
			if startDateStr != "" {
				startDate, err := time.Parse("2006-01-02T15:04", startDateStr)
				if err == nil {
					form.StartDate = &startDate
				}
			}

			endDateStr := r.FormValue("offer_end_date")
			if endDateStr != "" {
				endDate, err := time.Parse("2006-01-02T15:04", endDateStr)
				if err == nil {
					form.EndDate = &endDate
				}
			}

			if err := offers.CreateOffer(form); err != nil {
				if r.Header.Get("HX-Request") == "true" {
					tmpl, _ := template.ParseFiles("web/templates/admin-error-message.html")
					tmpl.Execute(w, err.Error())
				} else {
					http.Error(w, err.Error(), http.StatusBadRequest)
				}
				return
			}

			if r.Header.Get("HX-Request") == "true" {
				w.Header().Set("HX-Trigger", "refreshOffers")
				w.WriteHeader(http.StatusCreated)
				return
			}
			w.WriteHeader(http.StatusCreated)

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	// Admin offers products endpoint - returns available products for selection
	http.HandleFunc("/api/admin/offers/products", admin.RequireRole("admin", "product_admin")(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		products, err := offers.GetProductsForOfferSelection()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "text/html")
		for _, p := range products {
			fmt.Fprintf(w, `<option value="%d">%s</option>`, p.ID, p.Name)
		}
	}))

	// Admin offer detail/update/delete routes
	http.HandleFunc("/api/admin/offers/", admin.RequireRole("admin", "product_admin")(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimPrefix(r.URL.Path, "/api/admin/offers/")

		switch r.Method {
		case http.MethodPut:
			id, err := strconv.Atoi(path)
			if err != nil {
				http.Error(w, "Invalid offer ID", http.StatusBadRequest)
				return
			}

			if err := r.ParseForm(); err != nil {
				http.Error(w, "Invalid form data", http.StatusBadRequest)
				return
			}

			offerPrice, err := strconv.ParseFloat(r.FormValue("offer_price"), 64)
			if err != nil {
				http.Error(w, "Invalid offer price", http.StatusBadRequest)
				return
			}

			form := offers.OfferForm{
				OfferPrice: offerPrice,
			}

			// Parse optional dates
			startDateStr := r.FormValue("offer_start_date")
			if startDateStr != "" {
				startDate, err := time.Parse("2006-01-02T15:04", startDateStr)
				if err == nil {
					form.StartDate = &startDate
				}
			}

			endDateStr := r.FormValue("offer_end_date")
			if endDateStr != "" {
				endDate, err := time.Parse("2006-01-02T15:04", endDateStr)
				if err == nil {
					form.EndDate = &endDate
				}
			}

			if err := offers.UpdateOffer(id, form); err != nil {
				if r.Header.Get("HX-Request") == "true" {
					tmpl, _ := template.ParseFiles("web/templates/admin-error-message.html")
					tmpl.Execute(w, err.Error())
				} else {
					http.Error(w, err.Error(), http.StatusBadRequest)
				}
				return
			}

			if r.Header.Get("HX-Request") == "true" {
				w.Header().Set("HX-Trigger", "refreshOffers")
				w.WriteHeader(http.StatusOK)
				return
			}
			w.WriteHeader(http.StatusOK)

		case http.MethodDelete:
			id, err := strconv.Atoi(path)
			if err != nil {
				http.Error(w, "Invalid offer ID", http.StatusBadRequest)
				return
			}

			_, err = offers.ToggleOfferStatus(id)
			if err != nil {
				if r.Header.Get("HX-Request") == "true" {
					tmpl, _ := template.ParseFiles("web/templates/admin-error-message.html")
					tmpl.Execute(w, err.Error())
				} else {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
				return
			}

			if r.Header.Get("HX-Request") == "true" {
				w.Header().Set("HX-Trigger", "refreshOffers")
				w.WriteHeader(http.StatusOK)
				return
			}
			w.WriteHeader(http.StatusOK)

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
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
		funcMap := orderFuncMap()
		tmpl, err := template.New("admin-order-detail.html").Funcs(funcMap).ParseFiles("web/templates/admin-order-detail.html")
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

	http.HandleFunc("/api/admin/orders/{id}/status-modal", admin.RequireRole("admin", "product_admin")(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		path := strings.TrimPrefix(r.URL.Path, "/api/admin/orders/")
		path = strings.TrimSuffix(path, "/status-modal")
		id, err := strconv.Atoi(path)
		if err != nil {
			http.Error(w, "Invalid order ID", http.StatusBadRequest)
			return
		}

		order, err := orders.GetOrderByID(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		setCacheHeaders(w, 86400) // 1 day cache for static modal
		w.Header().Set("Content-Type", "text/html")
		tmpl, err := template.ParseFiles("web/templates/admin-order-status-modal.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, map[string]interface{}{
			"OrderID":       order.ID,
			"CurrentStatus": order.Status,
		})
	}))

	http.HandleFunc("/api/admin/orders/{id}/status", admin.RequireRole("admin", "product_admin")(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		path := strings.TrimPrefix(r.URL.Path, "/api/admin/orders/")
		path = strings.TrimSuffix(path, "/status")
		id, err := strconv.Atoi(path)
		if err != nil {
			http.Error(w, "Invalid order ID", http.StatusBadRequest)
			return
		}

		status := r.FormValue("status")
		if status == "" {
			http.Error(w, "Status is required", http.StatusBadRequest)
			return
		}

		err = orders.UpdateOrderStatus(id, status)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Return the updated order row
		order, err := orders.GetOrderByID(id)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		role, _ := admin.RoleFromRequest(r)
		canViewFinancialData := role == "admin"

		w.Header().Set("Content-Type", "text/html")
		funcMap := orderFuncMap()
		tmpl, err := template.New("admin-order-row.html").Funcs(funcMap).ParseFiles("web/templates/admin-order-row.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, map[string]interface{}{
			"Order":                order,
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
			description := r.FormValue("description")
			sku := r.FormValue("sku")
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

			isAvailable := isAvailableStr == "on"
			// Create product
			product, err := products.CreateProduct(name, price, category, description, sku, isAvailable, brandIDs, fitsProductIDs)
			if err != nil {
				// Return error message for HTMX
				if r.Header.Get("HX-Request") == "true" {
					tmpl, _ := template.ParseFiles("web/templates/admin-error-message.html")
					tmpl.Execute(w, err.Error())
				} else {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
				return
			}

			// Handle multiple image uploads
			if err := handleMultipleImageUploads(r, "images", product.ProductID); err != nil {
				// Clean up product if image upload fails
				products.DeleteProduct(product.ID)
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
		// Extract path after /api/admin/products/
		path := strings.TrimPrefix(r.URL.Path, "/api/admin/products/")
		parts := strings.Split(path, "/")

		// Check if this is an image delete request: /{productID}/images/{imageID}
		if len(parts) >= 3 && parts[1] == "images" {
			productID, err := strconv.Atoi(parts[0])
			if err != nil {
				http.Error(w, "Invalid product ID", http.StatusBadRequest)
				return
			}
			imageID, err := strconv.Atoi(parts[2])
			if err != nil {
				http.Error(w, "Invalid image ID", http.StatusBadRequest)
				return
			}

			if r.Method == http.MethodDelete {
				// Get image to find file path
				images, err := products.GetProductImages(productID)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				var imageURL string
				for _, img := range images {
					if img.ID == imageID {
						imageURL = img.ImageURL
						break
					}
				}

				// Delete from database
				if err := products.DeleteProductImage(imageID); err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}

				// Delete file from filesystem if it's in uploads folder
				if strings.Contains(imageURL, "/uploads/") {
					filePath := filepath.Join("web/static", strings.TrimPrefix(imageURL, "/static/"))
					os.Remove(filePath)
				}

				// Return empty response for HTMX
				if r.Header.Get("HX-Request") == "true" {
					w.WriteHeader(http.StatusOK)
				} else {
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(map[string]string{"message": "Image deleted successfully"})
				}
				return
			}
		}

		// Regular product ID parsing
		id, err := strconv.Atoi(path)
		if err != nil {
			http.Error(w, "Invalid product ID", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodPut:
			// Parse multipart form
			if err := r.ParseMultipartForm(maxUploadSize); err != nil {
				fmt.Printf("\n%v\n", err.Error())
				http.Error(w, "File too large or invalid form data", http.StatusBadRequest)
				return
			}

			// Get form values
			name := r.FormValue("name")
			priceStr := r.FormValue("price")
			category := r.FormValue("category")
			description := r.FormValue("description")
			sku := r.FormValue("sku")
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

			// Update product
			if err := products.UpdateProduct(id, name, price, category, description, sku, isAvailable, brandIDs, fitsProductIDs); err != nil {
				// Return error message for HTMX
				if r.Header.Get("HX-Request") == "true" {
					tmpl, _ := template.ParseFiles("web/templates/admin-error-message.html")
					tmpl.Execute(w, err.Error())
				} else {
					http.Error(w, err.Error(), http.StatusInternalServerError)
				}
				return
			}

			// Get product ID for images
			product, err := products.GetProductByID(id)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			// Handle multiple image uploads
			if err := handleMultipleImageUploads(r, "images", product.ProductID); err != nil {
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

			// Get product images
			productImages, err := products.GetProductImages(product.ProductID)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			tmpl, err := template.ParseFiles("web/templates/admin-edit-form.html")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			editData := adminEditData{
				Product:         product,
				ProductImages:   productImages,
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

func handleMultipleImageUploads(r *http.Request, fieldName string, productID int) error {
	// Parse multipart form if not already parsed
	if r.MultipartForm == nil {
		if err := r.ParseMultipartForm(maxUploadSize); err != nil {
			return fmt.Errorf("failed to parse form: %v", err)
		}
	}

	// Get the files
	files := r.MultipartForm.File[fieldName]
	if len(files) == 0 {
		return nil // No images to upload
	}

	// Process each uploaded file
	for i, fileHeader := range files {
		file, err := fileHeader.Open()
		if err != nil {
			return fmt.Errorf("failed to open file %s: %v", fileHeader.Filename, err)
		}
		defer file.Close()

		// Validate file size
		if fileHeader.Size > maxUploadSize {
			return fmt.Errorf("file %s exceeds maximum allowed size of 5MB", fileHeader.Filename)
		}

		// Validate file type
		contentType := fileHeader.Header.Get("Content-Type")
		if !strings.HasPrefix(contentType, "image/") {
			return fmt.Errorf("file %s must be an image", fileHeader.Filename)
		}

		// Generate unique filename
		ext := filepath.Ext(fileHeader.Filename)
		randomBytes := make([]byte, 16)
		if _, err := rand.Read(randomBytes); err != nil {
			return fmt.Errorf("failed to generate random filename: %v", err)
		}
		filename := hex.EncodeToString(randomBytes) + ext

		// Create file path
		filePath := filepath.Join(uploadPath, filename)

		// Create destination file
		dst, err := os.Create(filePath)
		if err != nil {
			return fmt.Errorf("failed to create file: %v", err)
		}
		defer dst.Close()

		// Copy uploaded file to destination
		if _, err := io.Copy(dst, file); err != nil {
			os.Remove(filePath)
			return fmt.Errorf("failed to save file: %v", err)
		}

		imagePath := "/static/images/uploads/" + filename
		isPrimary := i == 0 // First image is primary

		_, err = products.CreateProductImage(productID, imagePath, i, isPrimary)
		if err != nil {
			os.Remove(filePath)
			return fmt.Errorf("failed to save image to database: %v", err)
		}
	}

	return nil
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
