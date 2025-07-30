package rates

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	_ "github.com/lib/pq"
)

var database *sql.DB

func InitDataBaseInterface() error {
	const maxAttempts = 10

	dsn := "host=db port=5432 user=postgres password=postgres dbname=esr sslmode=disable"

	var err error
	for i := 1; i <= maxAttempts; i++ {
		database, err = sql.Open("postgres", dsn)
		if err != nil {
			return fmt.Errorf("Failed to open DB: %v", err)
		}

		err = database.Ping()
		if err == nil {
			fmt.Println("Connected to database!")
			return nil
		}
		time.Sleep(2 * time.Second)
	}
	return fmt.Errorf("Failed to connect to DB: %v", err)
}

func CloseDB() {
	if database != nil {
		database.Close()
	}
}

func parseCurrencyPair(input string) (string, string, error) {
	if len(input) != 7 {
		return "", "", fmt.Errorf("invalid currency pair format: expected 'XXX/YYY'")
	}
	upper := strings.ToUpper(input)
	parts := strings.Split(upper, "/")

	if len(parts) != 2 {
		return "", "", fmt.Errorf("invalid currency pair format: expected 'XXX/YYY'")
	}

	if len(parts[0]) != 3 || len(parts[1]) != 3 {
		return "", "", fmt.Errorf("invalid currency pair format: expected 'XXX/YYY'")
	}

	return parts[0], parts[1], nil
}

func placeRequest(CurrencyPairCode string) (uint64, error) {
	var id uint64

	cur1, cur2, err := parseCurrencyPair(CurrencyPairCode)
	if err != nil {
		return 0, err
	}

	if database == nil {
		return 0, fmt.Errorf("database is not initialized yet")
	}

	query := `
        INSERT INTO update_requests (currency1, currency2)
        VALUES ($1, $2)
        RETURNING id;
    `
	err = database.QueryRow(query, cur1, cur2).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to insert rate: %w", err)
	}

	return id, nil
}

func getRateByPair(CurrencyPairCode string) (float64, error) {
	return 0, errors.New("no such pair")
}

func getRateByRequestId(requestId uint64) (float64, error) {

	return float64(requestId) * 0.1, nil
}
