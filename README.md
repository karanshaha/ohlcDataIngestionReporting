# OHLC Data Ingestion & Reporting API

This service ingests OHLC data from CSV files and stores it in PostgreSQL. It exposes HTTP endpoints to upload and query data, and is fully containerized with Docker.

[![Docker](https://img.shields.io/badge/Docker-Ready-brightgreen.svg)](https://www.docker.com/)
[![Go](https://img.shields.io/badge/Go-1.22-blue.svg)](https://golang.org/)

---
Github Repo : https://github.com/karanshaha/ohlcDataIngestionReporting

---

## 🎯 Features

- **Fast CSV ingestion** with configurable batch size & workers
- **Production Docker setup** with Postgres
- **CORS enabled** for frontend integration
- **Structured logging** with performance metrics
- **Health checks** & monitoring ready

---

## Prerequisites
1.Docker installed and running
2.Docker Compose installed (if using docker-compose.yml)

## 🚀 Quick start/setup (5 seconds)

```bash
# Clone & start
git clonehttps://github.com/karanshaha/ohlcDataIngestionReporting.git
cd ohlcDataIngestionReporting
docker compose up --build
```

## Running with DockerEnvironment variables
```bash
# The API is configurable via environment variables:

1. DATABASE_DSN – Postgres connection string (inside Docker, usually points to the db container)
   DATABASE_DSN=postgres://ohlc:ohlc@db:5432/ohlc?sslmode=disable

2. INGEST_BATCH_SIZE – Number of records per batch insert (e.g. 5000)
   INGEST_BATCH_SIZE=10000
 
3. INGEST_WORKERS – Number of concurrent workers processing the CSV (e.g. 4)
   INGEST_WORKERS=8 
```

## Running with Docker Compose
Docker compose link : https://github.com/karanshaha/ohlcDataIngestionReporting/blob/master/docker-compose.yml
```bash
# Build images
docker compose build

# Start services (API + Postgres)
docker compose up 

OR 

#To run in the background
docker compose up -y -d

#To stop
docker compose down

````

## Directly pull the image and run the service published on Docker hub
Docker hub : https://hub.docker.com/r/karansshaha/ohlcdataingestionreporting-api


## 3. Run only the API container (if DB is external)
````
docker run --rm -p 8080:8080 \
  -e DATABASE_DSN="postgres://user:pass@host:5432/dbname?sslmode=disable" \
  -e INGEST_BATCH_SIZE=10000 \
  -e INGEST_WORKERS=8 \
  --name <image-name> \
````

## API endpoints
```bash
------------------------------------------------------------------------------------

# Health check
Req: curl --location --request GET 'http://localhost:8080/health'

Response : 
{
   "status":"ok"
}

------------------------------------------------------------------------------------

# POST /data
Req: curl -X POST http://localhost:8080/data -F "file=@sample.csv"

Response :
{
    "processed_rows": 20000
}

------------------------------------------------------------------------------------

# GET /data
Req : curl --location 'http://localhost:8080/data?symbol=BTCUSDT&limit=1&offset=0'

Response :
{
    "data": [
        {
            "id": 3,
            "unix": 1735689660000,
            "symbol": "BTCUSDT",
            "open": 42112.27827193,
            "high": 42121.37446253,
            "low": 42055.63119778,
            "close": 42055.92823561
        }
    ],
    "limit": 1,
    "offset": 0,
    "total_count": 18158,
    "has_more": true
}
------------------------------------------------------------------------------------
````

## ** CSV generator utility **
Now quickly generate test data with csv gen service using any row count via `NUM_ROWS` env var.[1] and is containersed as well
it will generate a csv file in the current directory and has volume mounted to ./test-data folder.


https://github.com/karanshaha/ohlcDataIngestionReporting/blob/master/generate_csv.go


## Run using docker compose or command 
Github link : https://github.com/karanshaha/ohlcDataIngestionReporting/blob/master/docker-compose-csv-remote.yml
````
 docker compose -f docker-compose-csv-remote.yml run --rm -e NUM_ROWS=200 csv-gen
 
 OR 
 
 docker pull karansshaha/ohlcdataingestionreporting-csv-gen:v1.0.0
 
 docker run --rm \
  -e NUM_ROWS=2 \
  -v /path/to/local/data:/app/test-data \
  karansshaha/ohlcdataingestionreporting-csv-gen:v1.0.0

````
Docker hub : https://hub.docker.com/r/karansshaha/ohlcdataingestionreporting-csv-gen

## Assumptions
1. Currently, allowing duplicate rows to be ingested 
2. Based on business requirement, we can add a unique constraint on symbol or data. 
3. In case of any exceptions current behavior is to log the error rollback the transaction and all or nothing

## Exception Handling or enhancements
1. Can certainly add more exception handling and logging and retry logic.
2. In case the file is corrupt or processing is taking too long or strolls can add async handling using kafka
3. Also file can be uploaded to S3 bucket and processed asynchronously in case fails processing can be retried using cron job.
