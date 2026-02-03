package main

import (
	"encoding/json"
	"fmt"
	"kasir-api/config"
	"kasir-api/handlers"
	"kasir-api/repositories"
	"kasir-api/services"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/viper"
)

type Category struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
	Description string `json:"description"`
}

var categories = []Category{
	{ID: 1, Name: "Makanan", Description: "Kategori makanan"},
	{ID: 2, Name: "Minuman", Description: "Kategori minuman"},
}

func getCategories(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(categories)
}

func createCategory(w http.ResponseWriter, r *http.Request) {
	var newCategory Category
	err := json.NewDecoder(r.Body).Decode(&newCategory)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	newCategory.ID = len(categories) + 1
	categories = append(categories, newCategory)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(newCategory)
}

func getCategoryByID(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/categories/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	for _, category := range categories {
		if category.ID == id {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(category)
			return
		}
	}

	http.Error(w, "Category not found", http.StatusNotFound)
}

func updateCategory(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/categories/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	var updatedCategory Category
	err = json.NewDecoder(r.Body).Decode(&updatedCategory)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	for i, category := range categories {
		if category.ID == id {
			updatedCategory.ID = id
			categories[i] = updatedCategory
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(updatedCategory)
			return
		}
	}

	http.Error(w, "Category not found", http.StatusNotFound)
}

func deleteCategory(w http.ResponseWriter, r *http.Request) {
	idStr := strings.TrimPrefix(r.URL.Path, "/api/categories/")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		http.Error(w, "Invalid category ID", http.StatusBadRequest)
		return
	}

	for i, category := range categories {
		if category.ID == id {
			categories = append(categories[:i], categories[i+1:]...)
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"message": "Category deleted successfully"})
			return
		}
	}

	http.Error(w, "Category not found", http.StatusNotFound)
}

type Config struct {
	Port string `mapstructure:"PORT"`
	DBConn string `mapstructure:"DB_CONN"`
}

func main() {
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if _, err := os.Stat(".env"); err == nil {
		viper.SetConfigFile(".env")
		_ = viper.ReadInConfig()
	}

	env := Config{
		Port: viper.GetString("PORT"),
		DBConn: viper.GetString("DB_CONN"),
	}
	
	//setup database
	db, err := config.InitDB(env.DBConn)
	if err != nil {
		log.Fatal("Failed to initialize database: ", err)
	}
	defer db.Close()

	productRepo := repositories.NewProductRepository(db)
	productService := services.NewProductService(productRepo)
	ProductHandler := handlers.NewProductHandler(productService)

	//setup router
	http.HandleFunc("api/products", ProductHandler.HandleProducts)
	http.HandleFunc("api/products/", ProductHandler.HandleProductByID)

	//GET /api/categories, POST /api/categories
	http.HandleFunc("/api/categories", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			getCategories(w, r)
		} else if r.Method == "POST" {
			createCategory(w, r)
		}
	})

	// GET /api/categories/{id}, PUT /api/categories/{id}, DELETE /api/categories/{id}
	http.HandleFunc("/api/categories/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			getCategoryByID(w, r)
		} else if r.Method == "PUT" {
			updateCategory(w, r)
		} else if r.Method == "DELETE" {
			deleteCategory(w, r)
		}
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status":  "OK",
			"message": "API Running",
		})
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		
		//Build endpoint metadata for current kasir API
		endpoints := map[string]map[string]map[string]string{
			"Products": {
				"list": {
					"method": "GET",
					"path":   "/api/products",
					"description": "List all products",
				},
				"create": {
					"method": "POST",
					"path":   "/api/products",
					"description": "Create a new product",
				},
				"get": {
					"method": "GET",
					"path":   "/api/products/{id}",
					"description": "Get a product by ID",
				},
				"update": {
					"method": "PUT",
					"path":   "/api/products/{id}",
					"description": "Update a product by ID",
				},
				"delete": {
					"method": "DELETE",
					"path":   "/api/products/{id}",
					"description": "Delete a product by ID",
				},
			},
			"Categories": {
				"list": {
					"method": "GET",
					"path":   "/api/categories",
					"description": "List all categories",
				},
				"create": {
					"method": "POST",
					"path":   "/api/categories",
					"description": "Create a new category",
				},
				"get": {
					"method": "GET",
					"path":   "/api/categories/{id}",
					"description": "Get a category by ID",
				},
				"update": {
					"method": "PUT",
					"path":   "/api/categories/{id}",
					"description": "Update a category by ID",
				},
				"delete": {
					"method": "DELETE",
					"path":   "/api/categories/{id}",
					"description": "Delete a category by ID",
				},
			},
			"Health": {
				"check": {
					"method": "GET",
					"path":   "/health",
					"description": "Check the health of the API",
				},
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]any{
			"status":  "OK",
			"message": "Kasir API is Running Successfully",
			"endpoints": endpoints,
		})
	})

	fmt.Println("Server running on localhost:" + env.Port)

	err = http.ListenAndServe(":" + env.Port, nil)
	if err != nil {
		fmt.Println(err)
	}
}
