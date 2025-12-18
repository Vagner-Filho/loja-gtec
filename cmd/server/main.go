package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
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
	"lojagtec/internal/database"
	"lojagtec/internal/orders"
	"lojagtec/internal/products"
)

const (
	maxUploadSize = 5 << 20 // 5MB
	uploadPath    = "web/static/images/uploads"
)

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
			PaymentMethod: r.FormValue("paymentMethod"),
			CardName:      r.FormValue("cardName"),
			CardNumber:    r.FormValue("cardNumber"),
			Expiry:        r.FormValue("expiry"),
			CVV:           r.FormValue("cvv"),
			CPF:           r.FormValue("cpf"),
			PixKey:        r.FormValue("pixKey"),
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
			tmpl.Execute(w, orders.ValidationError{Field: "general", Message: "Erro ao processar pedido: " + err.Error()})
			return
		}

		// TODO: Process payment with Stripe here
		// For now, we'll mark as paid for demo purposes
		err = orders.UpdateOrderPaymentStatus(order.ID, "paid", "demo_payment_id")
		if err != nil {
			log.Printf("Failed to update payment status: %v", err)
		}

		// Return success response with order details
		w.Header().Set("Content-Type", "text/html")
		tmpl, err := template.ParseFiles("web/templates/checkout-success.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Execute template with order data
		tmpl.Execute(w, map[string]interface{}{
			"Order": order,
		})
	})

	// Product filter routes
	http.HandleFunc("/products/bebedouros", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("web/templates/product-cards.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		prods, err := products.GetProductsByCategory("bebedouros")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		tmpl.Execute(w, prods)
	})

	http.HandleFunc("/products/purificadores", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("web/templates/product-cards.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		prods, err := products.GetProductsByCategory("purificadores")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, prods)
	})

	http.HandleFunc("/products/refis", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("web/templates/product-cards.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		prods, err := products.GetProductsByCategory("refis")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, prods)
	})

	http.HandleFunc("/products/pecas", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("web/templates/product-cards.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		prods, err := products.GetProductsByCategory("pecas")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, prods)
	})

	http.HandleFunc("/products/all", func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("web/templates/product-cards.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		prods, err := products.GetAllProducts()
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

	http.HandleFunc("/admin", admin.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		tmpl, err := template.ParseFiles("web/templates/admin-dashboard.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		tmpl.Execute(w, nil)
	}))

	// Admin API routes
	http.HandleFunc("/api/admin/products", admin.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
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

			if name == "" || priceStr == "" || category == "" {
				http.Error(w, "Missing required fields", http.StatusBadRequest)
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
			product, err := products.CreateProduct(name, price, imagePath, category, isAvailable)
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

	http.HandleFunc("/api/admin/products/", admin.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
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

			if name == "" || priceStr == "" || category == "" {
				http.Error(w, "Missing required fields", http.StatusBadRequest)
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
			if err := products.UpdateProduct(id, name, price, imagePath, category, isAvailable); err != nil {
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
	http.HandleFunc("/admin/products/", admin.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
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

			tmpl, err := template.ParseFiles("web/templates/admin-edit-form.html")
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			tmpl.Execute(w, product)
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
