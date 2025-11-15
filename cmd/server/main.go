package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"

	"lojagtec/internal/admin"
	"lojagtec/internal/database"
	"lojagtec/internal/products"
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

	fs := http.FileServer(http.Dir("web/static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

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

		prods, err := products.GetProductsByCategory("")
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
		w.Header().Set("Content-Type", "application/json")

		switch r.Method {
		case http.MethodGet:
			prods, err := products.GetAllProducts()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			json.NewEncoder(w).Encode(prods)

		case http.MethodPost:
			var req struct {
				Name     string  `json:"name"`
				Price    float64 `json:"price"`
				Image    string  `json:"image"`
				Category string  `json:"category"`
			}

			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			product, err := products.CreateProduct(req.Name, req.Price, req.Image, req.Category)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusCreated)
			json.NewEncoder(w).Encode(product)

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	http.HandleFunc("/api/admin/products/", admin.RequireAuth(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Extract ID from path
		path := strings.TrimPrefix(r.URL.Path, "/api/admin/products/")
		id, err := strconv.Atoi(path)
		if err != nil {
			http.Error(w, "Invalid product ID", http.StatusBadRequest)
			return
		}

		switch r.Method {
		case http.MethodPut:
			var req struct {
				Name     string  `json:"name"`
				Price    float64 `json:"price"`
				Image    string  `json:"image"`
				Category string  `json:"category"`
			}

			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			if err := products.UpdateProduct(id, req.Name, req.Price, req.Image, req.Category); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			json.NewEncoder(w).Encode(map[string]string{"message": "Product updated successfully"})

		case http.MethodDelete:
			if err := products.DeleteProduct(id); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			json.NewEncoder(w).Encode(map[string]string{"message": "Product deleted successfully"})

		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}))

	fmt.Println("Server starting at port 8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("Error starting server: %s\n", err)
	}
}
