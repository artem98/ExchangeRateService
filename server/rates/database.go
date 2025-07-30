package rates

import "errors"

var req uint64 = 0

func placeRequest(CurrencyPairCode string) (uint64, error) {
	req++
	return req, nil
}

func getRateByPair(CurrencyPairCode string) (float64, error) {
	return 0, errors.New("no such pair")
}

func getRateByRequestId(requestId uint64) (float64, error) {

	return float64(requestId) * 0.1, nil
}
