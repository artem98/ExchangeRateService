package db

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/artem98/ExchangeRateService/server/rates/external"
	_ "github.com/lib/pq"
)

type DataBase interface {
	GetRateByPair(cur1, cur2 string) (float64, time.Time, error)
	GetRateByRequestId(id uint64) (float64, time.Time, error)
	PlaceRequest(cur1, cur2 string) (uint64, error)
	MarkRequestAsProcessed(requestId uint64) error
	MarkRequestAsFailed(requestId uint64) error
	UpdateRate(currency1, currency2 string, rate float64) error
}

type DataBaseAdapter struct {
	database *sql.DB
}

func MakeDataBaseAdapter() (DataBaseAdapter, error) {
	db, err := initDataBaseInterface()
	if err != nil {
		return DataBaseAdapter{}, err
	}
	a := DataBaseAdapter{database: db}
	err = a.fillRatesAtStart()
	if err != nil {
		return DataBaseAdapter{}, fmt.Errorf("failed to fill DB: %v", err)
	}

	return a, nil
}

func initDataBaseInterface() (database *sql.DB, err error) {
	const maxAttempts = 10

	dsn := "host=db port=5432 user=postgres password=postgres dbname=esr sslmode=disable"

	for i := 1; i <= maxAttempts; i++ {
		database, err = sql.Open("postgres", dsn)
		if err != nil {
			return nil, fmt.Errorf("failed to open DB: %v", err)
		}

		err = database.Ping()
		if err == nil {
			fmt.Println("Connected to database!")
			break
		}
		time.Sleep(2 * time.Second)

	}
	if err != nil {
		return nil, fmt.Errorf("failed to connect to DB: %v", err)
	}

	return database, nil
}

func (a DataBaseAdapter) CloseDB() {
	if a.database != nil {
		a.database.Close()
	}
}

func (a DataBaseAdapter) fillRatesAtStart() error {
	if a.database == nil {
		return fmt.Errorf("database not initialized")
	}

	rows, err := a.database.Query("SELECT currency1, currency2, rate FROM rates")
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
		err = a.UpdateRate(currency1, currency2, rate)

		if err != nil {
			return err
		}
	}

	fmt.Println("Finished filling DB...")
	return nil
}

func (a DataBaseAdapter) PlaceRequest(currency1, currency2 string) (uint64, error) {
	var id uint64

	if a.database == nil {
		return 0, fmt.Errorf("database is not initialized yet")
	}

	query := `
        INSERT INTO update_requests (currency1, currency2, request_status)
        VALUES ($1, $2, $3)
        RETURNING id;
    `
	err := a.database.QueryRow(query, currency1, currency2, "submitted").Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to insert rate: %w", err)
	}

	return id, nil
}

func (a DataBaseAdapter) GetRateByPair(currency1, currency2 string) (float64, time.Time, error) {
	if a.database == nil {
		return 0, time.Time{}, errors.New("database not initialized")
	}

	var rate float64
	var timestamp time.Time
	query := `
        SELECT rate, update_time FROM rates
        WHERE currency1 = $1 AND currency2 = $2
        LIMIT 1
    `
	err := a.database.QueryRow(query, currency1, currency2).Scan(&rate, &timestamp)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, time.Time{}, errors.New("no such pair")
		}
		return 0, time.Time{}, fmt.Errorf("db query error: %w", err)
	}

	return rate, timestamp, nil
}

func (a DataBaseAdapter) GetRateByRequestId(requestId uint64) (float64, time.Time, error) {
	if a.database == nil {
		return 0, time.Time{}, errors.New("database not initialized")
	}

	var currency1, currency2 string
	err := a.database.QueryRow(`
        SELECT currency1, currency2 FROM update_requests
        WHERE id = $1
    `, requestId).Scan(&currency1, &currency2)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, time.Time{}, fmt.Errorf("no request with id %d", requestId)
		}
		return 0, time.Time{}, fmt.Errorf("failed to query request: %w", err)
	}

	return a.GetRateByPair(currency1, currency2)
}

func (a DataBaseAdapter) MarkRequestAsProcessed(requestId uint64) error {
	if a.database == nil {
		return fmt.Errorf("database not initialized")
	}

	query := `UPDATE update_requests SET request_status = 'ok' WHERE id = $1`
	_, err := a.database.Exec(query, requestId)
	if err != nil {
		return fmt.Errorf("failed to mark request as processed: %w", err)
	}
	return nil
}

func (a DataBaseAdapter) MarkRequestAsFailed(requestId uint64) error {
	if a.database == nil {
		return fmt.Errorf("database not initialized")
	}

	query := `UPDATE update_requests SET request_status = 'failed' WHERE id = $1`
	_, err := a.database.Exec(query, requestId)
	if err != nil {
		return fmt.Errorf("failed to mark request as failed: %w", err)
	}
	return nil
}

func (a DataBaseAdapter) UpdateRate(currency1, currency2 string, rate float64) error {
	if a.database == nil {
		return fmt.Errorf("database not initialized")
	}

	query := `
        INSERT INTO rates (currency1, currency2, rate, update_time)
        VALUES ($1, $2, $3, $4)
        ON CONFLICT (currency1, currency2) DO UPDATE
        SET rate = EXCLUDED.rate,
            update_time = EXCLUDED.update_time;
    `

	_, err := a.database.Exec(query, currency1, currency2, rate, time.Now())
	if err != nil {
		return fmt.Errorf("failed to upsert rate: %w", err)
	}

	return nil
}
