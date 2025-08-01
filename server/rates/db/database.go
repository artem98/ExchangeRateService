package db

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/artem98/ExchangeRateService/server/rates/external"
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

		rate, err := external.FetchRate(currency1, currency2)
		if err != nil {
			return err
		}
		err = UpdateRate(currency1, currency2, rate)

		if err != nil {
			return err
		}
	}

	fmt.Println("Finished filling DB...")
	return nil
}

func PlaceRequest(currency1, currency2 string) (uint64, error) {
	var id uint64

	if database == nil {
		return 0, fmt.Errorf("database is not initialized yet")
	}

	query := `
        INSERT INTO update_requests (currency1, currency2, request_status)
        VALUES ($1, $2, $3)
        RETURNING id;
    `
	err := database.QueryRow(query, currency1, currency2, "submitted").Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to insert rate: %w", err)
	}

	return id, nil
}

func GetRateByPair(currency1, currency2 string) (float64, time.Time, error) {
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

func GetRateByRequestId(requestId uint64) (float64, time.Time, error) {
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

	return GetRateByPair(currency1, currency2)
}

func MarkRequestAsProcessed(requestId uint64) error {
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

func MarkRequestAsFailed(requestId uint64) error {
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

func UpdateRate(currency1, currency2 string, rate float64) error {
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
