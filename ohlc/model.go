package ohlc

type Record struct {
	ID     int64   `json:"id"`
	UnixMS int64   `json:"unix"`
	Symbol string  `json:"symbol"`
	Open   float64 `json:"open"`
	High   float64 `json:"high"`
	Low    float64 `json:"low"`
	Close  float64 `json:"close"`
}
