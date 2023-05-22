package util

const (
	USD = "USD"
	EUR = "EUR"
	CAD = "CAD"
)

func isSupportedCurrency(currency string) bool {
	switch currency {
	case USD, EUR, CAD:
		return true
	}
	return false
}