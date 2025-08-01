package worker

import (
	"fmt"

	"github.com/artem98/ExchangeRateService/server/rates/db"
	"github.com/artem98/ExchangeRateService/server/rates/external"
)

func MakeRateUpdateJob(currency1, currency2 string, reqId uint64, db db.DataBase) Job {
	return func() (err error) {
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("Recovered in processJob:", r)
				db.MarkRequestAsFailed(reqId)
				err = fmt.Errorf("panic occurred: %v", r)
			}
		}()

		rate, err := external.FetchRate(currency1, currency2)

		if err != nil {
			db.MarkRequestAsFailed(reqId)
			return err
		}

		err = db.UpdateRate(currency1, currency2, rate)
		if err != nil {
			db.MarkRequestAsFailed(reqId)
			return err
		}

		return db.MarkRequestAsProcessed(reqId)
	}
}
