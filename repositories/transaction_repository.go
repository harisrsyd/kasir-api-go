package repositories

import (
	"database/sql"
	"fmt"
	"kasir-api/models"
	"strings"
	"time"
)

type TransactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (repo *TransactionRepository) CreateTransaction(items []models.CheckoutItem) (*models.Transaction, error) {
	tx, err := repo.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	totalAmount := 0
	details := make([]models.TransactionDetail, 0)

	for _, item := range items {
		var productPrice, stock int
		var productName string

		err := tx.QueryRow("SELECT name, price, stock FROM products WHERE id = $1", item.ProductID).Scan(&productName, &productPrice, &stock)
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("product id %d not found", item.ProductID)
		}
		if err != nil {
			return nil, err
		}

		subtotal := productPrice * item.Quantity
		totalAmount += subtotal

		_, err = tx.Exec("UPDATE products SET stock = stock - $1 WHERE id = $2", item.Quantity, item.ProductID)
		if err != nil {
			return nil, err
		}

		details = append(details, models.TransactionDetail{
			ProductID:   item.ProductID,
			ProductName: productName,
			Quantity:    item.Quantity,
			Subtotal:    subtotal,
		})
	}

	var transactionID int
	var createdAt time.Time
	err = tx.QueryRow("INSERT INTO transactions (total_amount) VALUES ($1) RETURNING id, created_at", totalAmount).Scan(&transactionID, &createdAt)
	if err != nil {
		return nil, err
	}

	query := `INSERT INTO transaction_details (transaction_id, product_id, quantity, subtotal) VALUES `

	args := []interface{}{}
	placeholders := []string{}

	for i, d := range details {
		base := i * 4
		placeholders = append(
			placeholders,
			fmt.Sprintf("($%d, $%d, $%d, $%d)", base+1, base+2, base+3, base+4),
		)

		args = append(args,
			transactionID,
			d.ProductID,
			d.Quantity,
			d.Subtotal,
		)
	}

	query += strings.Join(placeholders, ", ")

	_, err = tx.Exec(query, args...)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return &models.Transaction{
		ID:          transactionID,
		TotalAmount: totalAmount,
		CreatedAt:   createdAt,
		Details:     details,
	}, nil
}

func (repo *TransactionRepository) GetReportToday() (*models.Report, error) {
	var report models.Report
	err := repo.db.QueryRow("SELECT COUNT(id) AS total_sales, SUM(total_amount) AS total_revenue FROM transactions WHERE DATE(created_at) = DATE(NOW())").Scan(&report.TotalSales, &report.TotalRevenue)
	if err != nil {
		return nil, err
	}
	return &report, nil
}
