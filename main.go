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
	"strings"

	"github.com/spf13/viper"
)

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
	productHandler := handlers.NewProductHandler(productService)

	categoryRepo := repositories.NewCategoryRepository(db)
	categoryService := services.NewCategoryService(categoryRepo)
	categoryHandler := handlers.NewCategoryHandler(categoryService)

	transactionRepo := repositories.NewTransactionRepository(db)
	transactionService := services.NewTransactionService(transactionRepo)
	transactionHandler := handlers.NewTransactionHandler(transactionService)

	//setup router
	http.HandleFunc("/api/products", productHandler.HandleProducts)
	http.HandleFunc("/api/products/", productHandler.HandleProductByID)
	http.HandleFunc("/api/categories", categoryHandler.HandleCategories)
	http.HandleFunc("/api/categories/", categoryHandler.HandleCategoryByID)
	http.HandleFunc("/api/checkout", transactionHandler.HandleCheckout)
	http.HandleFunc("/api/report/today", transactionHandler.HandleReport)


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
