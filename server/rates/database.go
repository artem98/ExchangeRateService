package rates

import "errors"

func placeRequest() (uint64, error) {
	return 0, nil
}

func getRateByPair(CurrencyPairCode string) (float64, error) {
	return 0, errors.New("no such pair")
}

func getRateByRequestId(requestId uint64) (float64, error) {

	return float64(requestId) * 0.1, nil
}
