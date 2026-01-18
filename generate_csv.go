package main

import (
	"encoding/csv"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"
)

func main() {
	rand.Seed(time.Now().UnixNano())

	err := os.MkdirAll("/app/test_data", 0777)
	if err != nil {
		panic(err)
	}

	numRowsStr := os.Getenv("NUM_ROWS")
	numRows, err := strconv.Atoi(strings.TrimSpace(numRowsStr))
	if err != nil || numRows <= 0 {
		numRows = 10000 // Default
	}

	// Generate 10,000 rows of BTCUSDT + ETHUSDT data
	generateOHLC("/app/test_data/test_ohlc.csv", numRows)
	fmt.Println("Generated test_ohlc.csv")
}

func generateOHLC(filename string, numRows int) {
	f, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer f.Close()

	writer := csv.NewWriter(f)
	defer writer.Flush()

	// Write header (exact format from assignment)
	header := []string{"UNIX", "SYMBOL", "OPEN", "HIGH", "LOW", "CLOSE"}
	if err := writer.Write(header); err != nil {
		panic(err)
	}

	baseTime := time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC).UnixMilli()
	symbols := []string{"BTCUSDT", "ETHUSDT", "ABC", "DEF", "TEST"}

	for i := 0; i < numRows; i++ {
		symbol := symbols[rand.Intn(len(symbols))]

		// Realistic price movement
		basePrice := 42000.0
		if symbol == "ETHUSDT" {
			basePrice = 2500.0
		}
		if symbol == "TEST" {
			basePrice = 2000.0
		}

		price := basePrice + rand.Float64()*2000 - 1000 // ±$1000 fluctuation
		open := price
		high := open + rand.Float64()*100
		low := open - rand.Float64()*100
		close := low + rand.Float64()*(high-low)

		// Format exactly like assignment (13-digit UNIX ms, 8 decimals)
		unix := baseTime + int64(i*60000) // 1 minute candles
		row := []string{
			strconv.FormatInt(unix, 10),
			symbol,
			fmt.Sprintf("%.8f", open),
			fmt.Sprintf("%.8f", high),
			fmt.Sprintf("%.8f", low),
			fmt.Sprintf("%.8f", close),
		}

		fmt.Println("Done writing the data ")

		if err := writer.Write(row); err != nil {
			panic(err)
		}
	}
}
