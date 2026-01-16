package http

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"ohlcDataIngestionReporting/ohlc"
	"strconv"
	"sync"
)

type Handlers struct {
	Repo ohlc.Repository
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func validateHeader(row []string) bool {
	expected := []string{"UNIX", "SYMBOL", "OPEN", "HIGH", "LOW", "CLOSE"}
	if len(row) != len(expected) {
		return false
	}
	for i, v := range expected {
		if row[i] != v {
			return false
		}
	}
	return true
}

// @Summary Upload CSV data
// @Description Upload and process CSV for OHLC data
// @Tags data
// @Accept multipart/form-data
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 400 {string} error
// @Router /data [post]
func (h *Handlers) UploadData(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	file, _, err := r.FormFile("file")
	if err != nil {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "file is required"})
		return
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.ReuseRecord = true

	headerRow, err := reader.Read()
	if err != nil || !validateHeader(headerRow) {
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid header"})
		return
	}

	recordsCh := make(chan ohlc.Record, 10000)
	errCh := make(chan error, 1)

	const workerCount = 4
	const batchSize = 5000

	var wg sync.WaitGroup
	wg.Add(workerCount)

	for i := 0; i < workerCount; i++ {
		go func() {
			defer wg.Done()
			batch := make([]ohlc.Record, 0, batchSize)

			for rec := range recordsCh {
				batch = append(batch, rec)
				if len(batch) >= batchSize {
					if err := h.Repo.BulkInsert(ctx, batch); err != nil {
						select {
						case errCh <- err:
						default:
						}
						return
					}
					batch = batch[:0]
				}
			}

			if len(batch) > 0 {
				if err := h.Repo.BulkInsert(ctx, batch); err != nil {
					select {
					case errCh <- err:
					default:
					}
					return
				}
			}
		}()
	}

	var processed int64

producerLoop:
	for {
		row, err := reader.Read()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			errCh <- fmt.Errorf("invalid csv row: %w", err)
			break producerLoop
		}
		if len(row) != 6 {
			errCh <- fmt.Errorf("invalid column count")
			break producerLoop
		}

		unix, err := strconv.ParseInt(row[0], 10, 64)
		if err != nil {
			errCh <- fmt.Errorf("invalid unix: %w", err)
			break producerLoop
		}
		open, err1 := strconv.ParseFloat(row[2], 64)
		high, err2 := strconv.ParseFloat(row[3], 64)
		low, err3 := strconv.ParseFloat(row[4], 64)
		closeV, err4 := strconv.ParseFloat(row[5], 64)
		if err1 != nil || err2 != nil || err3 != nil || err4 != nil {
			errCh <- fmt.Errorf("invalid price")
			break producerLoop
		}

		processed++
		rec := ohlc.Record{
			UnixMS: unix,
			Symbol: row[1],
			Open:   open,
			High:   high,
			Low:    low,
			Close:  closeV,
		}

		select {
		case recordsCh <- rec:
		case <-ctx.Done():
			errCh <- ctx.Err()
			break producerLoop
		}
	}

	close(recordsCh)

	wgDone := make(chan struct{})
	go func() {
		wg.Wait()
		close(wgDone)
	}()

	select {
	case err := <-errCh:
		<-wgDone
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	case <-wgDone:
	case <-ctx.Done():
		writeJSON(w, http.StatusRequestTimeout, map[string]string{"error": "request cancelled"})
		return
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"processed_rows": processed,
	})
}

// PaginationResponse godoc
// @Description Pagination metadata and data
type PaginationResponse struct {
	Data    []ohlc.Record `json:"data"`
	Limit   int           `json:"limit"`
	Offset  int           `json:"offset"`
	Total   int64         `json:"total_count"`
	HasMore bool          `json:"has_more"`
}

// @Summary Query data with pagination
// @Description List OHLC data with optional pagination
// @Tags data
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param limit query int false "Items per page" default(10)
// @Success 200 {object} PaginationResponse
// @Router /data [get]
func (h *Handlers) ListData(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()
	q := r.URL.Query()

	symbol := q.Get("symbol")
	limit, offset := parsePagination(q.Get("limit"), q.Get("offset"))

	records, total, err := h.Repo.QueryPaginated(ctx, symbol, limit, offset)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "query failed"})
		return
	}

	hasMore := int64(offset+limit) < total

	writeJSON(w, http.StatusOK, PaginationResponse{
		Data:    records,
		Limit:   limit,
		Offset:  offset,
		Total:   total,
		HasMore: hasMore,
	})
}

func parsePagination(limitStr, offsetStr string) (limit, offset int) {
	limit = 100
	offset = 0

	if limitStr != "" {
		if parsed, err := strconv.Atoi(limitStr); err == nil && parsed > 0 && parsed <= 1000 {
			limit = parsed
		}
	}
	if offsetStr != "" {
		if parsed, err := strconv.Atoi(offsetStr); err == nil && parsed >= 0 {
			offset = parsed
		}
	}
	return
}
