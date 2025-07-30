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
			return fmt.Errorf("failed to open DB: %v", err)
		}

		err = database.Ping()
		if err == nil {
			fmt.Println("Connected to database!")
			break
		}
		time.Sleep(2 * time.Second)

	}
	if err != nil {
		return fmt.Errorf("failed to connect to DB: %v", err)
	}
	err = fillRatesAtStart()
	if err != nil {
		return fmt.Errorf("failed to fill DB: %v", err)
	}
	return nil
}

func CloseDB() {
	if database != nil {
		database.Close()
	}
}

func fillRatesAtStart() error {
	if database == nil {
		return fmt.Errorf("database not initialized")
	}

	rows, err := database.Query("SELECT currency1, currency2, rate FROM rates")
	if err != nil {
		return fmt.Errorf("failed to fetch currency pairs: %w", err)
	}
	defer rows.Close()

	fmt.Println("Start filling DB...")
	for rows.Next() {
		var currency1, currency2 string
		var tableRate sql.NullFloat64
		err := rows.Scan(&currency1, &currency2, &tableRate)
		if err != nil {
			return fmt.Errorf("failed to scan row: %w", err)
		}

		if tableRate.Valid {
			continue
		}

		rate, err := fetchRate(currency1, currency2)
		if err != nil {
			return err
		}

		err = updateRate(currency1, currency2, rate)

		if err != nil {
			return err
		}
	}

	fmt.Println("Finished filling DB...")
	return nil
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

	currency1, currency2, err := parseCurrencyPair(CurrencyPairCode)
	if err != nil {
		return 0, err
	}

	if database == nil {
		return 0, fmt.Errorf("database is not initialized yet")
	}

	query := `
        INSERT INTO update_requests (currency1, currency2, request_status)
        VALUES ($1, $2, $3)
        RETURNING id;
    `
	err = database.QueryRow(query, currency1, currency2, "submitted").Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to insert rate: %w", err)
	}

	PlanJob(Job{Currency1: currency1, Currency2: currency2, reqId: id})

	return id, nil
}

func getRateByPairCode(CurrencyPairCode string) (float64, time.Time, error) {
	currency1, currency2, err := parseCurrencyPair(CurrencyPairCode)
	if err != nil {
		return 0, time.Time{}, err
	}

	return getRateByPair(currency1, currency2)
}

func getRateByPair(currency1, currency2 string) (float64, time.Time, error) {
	if database == nil {
		return 0, time.Time{}, errors.New("database not initialized")
	}

	var rate float64
	var timestamp time.Time
	query := `
        SELECT rate, update_time FROM rates
        WHERE currency1 = $1 AND currency2 = $2
        LIMIT 1
    `
	err := database.QueryRow(query, currency1, currency2).Scan(&rate, &timestamp)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, time.Time{}, errors.New("no such pair")
		}
		return 0, time.Time{}, fmt.Errorf("db query error: %w", err)
	}

	return rate, timestamp, nil
}

func getRateByRequestId(requestId uint64) (float64, time.Time, error) {
	if database == nil {
		return 0, time.Time{}, errors.New("database not initialized")
	}

	var currency1, currency2 string
	err := database.QueryRow(`
        SELECT currency1, currency2 FROM update_requests
        WHERE id = $1
    `, requestId).Scan(&currency1, &currency2)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, time.Time{}, fmt.Errorf("no request with id %d", requestId)
		}
		return 0, time.Time{}, fmt.Errorf("failed to query request: %w", err)
	}

	return getRateByPair(currency1, currency2)
}

func markRequestAsProcessed(requestId uint64) error {
	if database == nil {
		return fmt.Errorf("database not initialized")
	}

	query := `UPDATE update_requests SET request_status = 'ok' WHERE id = $1`
	_, err := database.Exec(query, requestId)
	if err != nil {
		return fmt.Errorf("failed to mark request as processed: %w", err)
	}
	return nil
}

func markRequestAsFailed(requestId uint64) error {
	if database == nil {
		return fmt.Errorf("database not initialized")
	}

	query := `UPDATE update_requests SET request_status = 'failed' WHERE id = $1`
	_, err := database.Exec(query, requestId)
	if err != nil {
		return fmt.Errorf("failed to mark request as failed: %w", err)
	}
	return nil
}

func updateRate(currency1, currency2 string, rate float64) error {
	if database == nil {
		return fmt.Errorf("database not initialized")
	}

	query := `
        INSERT INTO rates (currency1, currency2, rate, update_time)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (currency1, currency2) DO UPDATE
        SET rate = EXCLUDED.rate,
            update_time = EXCLUDED.update_time;
    `

	_, err := database.Exec(query, currency1, currency2, rate, time.Now())
	if err != nil {
		return fmt.Errorf("failed to upsert rate: %w", err)
	}

	return nil
}
