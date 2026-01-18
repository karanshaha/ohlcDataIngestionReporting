package http_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	ohttp "ohlcDataIngestionReporting/http"
	"ohlcDataIngestionReporting/ohlc"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRepository struct {
	mock.Mock
}

// Implement interface methods
func (m *MockRepository) BulkInsert(ctx context.Context, records []ohlc.Record) error {
	args := m.Called(ctx, records)
	return args.Error(0)
}

func (m *MockRepository) QueryPaginated(ctx context.Context, symbol string, limit, offset int) ([]ohlc.Record, int64, error) {
	args := m.Called(ctx, symbol, limit, offset)
	return args.Get(0).([]ohlc.Record), 2, args.Error(2)
}

func TestUploadData_ValidCSV(t *testing.T) {
	// Arrange
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, _ := writer.CreateFormFile("file", "test.csv")
	csv := "UNIX,SYMBOL,OPEN,HIGH,LOW,CLOSE\n" +
		"1735689600000,BTCUSDT,42123.29,42148.32,42120.82,42146.06\n"
	part.Write([]byte(csv))
	writer.Close()

	mockRepo := &MockRepository{}
	mockRepo.On("BulkInsert", mock.Anything, mock.MatchedBy(func(recs []ohlc.Record) bool {
		return len(recs) > 0
	})).Return(nil)

	h := &ohttp.Handlers{
		Repo:        mockRepo,
		BatchSize:   100, // small for test (1 batch)
		WorkerCount: 1,   // single worker for test
	}
	req := httptest.NewRequest(http.MethodPost, "/data", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rr := httptest.NewRecorder()

	// Act
	h.UploadData(rr, req)

	// Assert
	assert.Equal(t, http.StatusCreated, rr.Code, "expected 201")
	assert.True(t, mockRepo.AssertNumberOfCalls(t, "BulkInsert", 1))
}

func TestUploadData_ValidationError(t *testing.T) {
	// Arrange
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, _ := writer.CreateFormFile("file", "test.csv")
	csv := "UNIX,SYMBOL,OPEN,HIGH,LOW\n" +
		"1735689600000,BTCUSDT,42123.29,42148.32,42120.82,\n"
	part.Write([]byte(csv))
	writer.Close()

	mockRepo := &MockRepository{}
	mockRepo.On("BulkInsert", mock.Anything, mock.MatchedBy(func(recs []ohlc.Record) bool {
		return len(recs) > 0
	})).Return(nil)

	h := &ohttp.Handlers{
		Repo:        mockRepo,
		BatchSize:   100, // small for test (1 batch)
		WorkerCount: 1,   // single worker for test
	}
	req := httptest.NewRequest(http.MethodPost, "/data", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	rr := httptest.NewRecorder()

	// Act
	h.UploadData(rr, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, rr.Code, "expected 400")

	var got map[string]string
	json.Unmarshal(rr.Body.Bytes(), &got)

	assert.Equal(t, map[string]string{"error": "invalid header"}, got)
	assert.True(t, mockRepo.AssertNumberOfCalls(t, "BulkInsert", 0))
}

func TestListData_PaginationAndFilter(t *testing.T) {
	// Arrange: mock repo returns BTC data
	mockRepo := &MockRepository{}

	// Sample records
	records := []ohlc.Record{
		{ID: 1, Symbol: "BTCUSDT", Close: 42146.06},
		{ID: 2, Symbol: "BTCUSDT", Close: 42300.00},
	}

	mockRepo.On("QueryPaginated",
		mock.Anything,
		"BTCUSDT",
		10,
		0,
	).Return(records, int64(2), nil).Once()

	h := &ohttp.Handlers{Repo: mockRepo}
	req := httptest.NewRequest(http.MethodGet,
		"/data?symbol=BTCUSDT&limit=10&offset=0", nil)
	rr := httptest.NewRecorder()

	// Act
	h.ListData(rr, req)

	// Assert: status + JSON structure
	assert.Equal(t, http.StatusOK, rr.Code)

	var resp ohttp.PaginationResponse
	assert.NoError(t, json.Unmarshal(rr.Body.Bytes(), &resp))
	fmt.Print(resp)
	assert.Len(t, resp.Data, 2)
	assert.Equal(t, int64(2), resp.Total)
	assert.False(t, resp.HasMore)
	assert.Equal(t, "BTCUSDT", resp.Data[0].Symbol)

	// Verify exact repo call
	assert.True(t, mockRepo.AssertExpectations(t))
	assert.True(t, mockRepo.AssertNumberOfCalls(t, "QueryPaginated", 1))
}
