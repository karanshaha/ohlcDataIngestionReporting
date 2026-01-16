CREATE TABLE IF NOT EXISTS ohlc (
   id          BIGSERIAL PRIMARY KEY,
   ts_unix_ms  BIGINT        NOT NULL,
   symbol      TEXT          NOT NULL,
   open        NUMERIC(20,8) NOT NULL,
    high        NUMERIC(20,8) NOT NULL,
    low         NUMERIC(20,8) NOT NULL,
    close       NUMERIC(20,8) NOT NULL
    );

CREATE INDEX IF NOT EXISTS idx_ohlc_symbol_ts ON ohlc(symbol, ts_unix_ms);
