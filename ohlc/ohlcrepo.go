package ohlc

import (
	"context"
	"database/sql"
)

type Repository interface {
	BulkInsert(ctx context.Context, records []Record) error
	QueryPaginated(ctx context.Context, symbol string, limit, offset int) ([]Record, int64, error)
}

type repo struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) Repository {
	return &repo{db: db}
}

func (r *repo) BulkInsert(ctx context.Context, records []Record) error {
	if len(records) == 0 {
		return nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	stmt, err := tx.PrepareContext(ctx, `
		INSERT INTO ohlc (ts_unix_ms, symbol, open, high, low, close)
		VALUES ($1,$2,$3,$4,$5,$6)
	`)
	if err != nil {
		_ = tx.Rollback()
		return err
	}
	defer stmt.Close()

	for _, rec := range records {
		if _, err := stmt.ExecContext(ctx,
			rec.UnixMS, rec.Symbol, rec.Open, rec.High, rec.Low, rec.Close,
		); err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (r *repo) QueryPaginated(ctx context.Context, symbol string, limit, offset int) ([]Record, int64, error) {
	var total int64
	if err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*) FROM ohlc
		WHERE ($1 = '' OR symbol = $1)
	`, symbol).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, ts_unix_ms, symbol, open, high, low, close
		FROM ohlc
		WHERE ($1 = '' OR symbol = $1)
		ORDER BY ts_unix_ms
		LIMIT $2 OFFSET $3
	`, symbol, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var out []Record
	for rows.Next() {
		var rec Record
		if err := rows.Scan(&rec.ID, &rec.UnixMS, &rec.Symbol, &rec.Open, &rec.High, &rec.Low, &rec.Close); err != nil {
			return nil, 0, err
		}
		out = append(out, rec)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	return out, total, nil
}
