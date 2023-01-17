# sample-golang-app-s3
Golang app to demonstrate IAM role authentication to s3 resources

## Build Docker image

```bash
docker build -t sample-golang-app:latest .
```

# Run (local, from src)
```
# exmple 1
export S3_CONFIG__REGION="***" && \
export S3_CONFIG__BUCKET_NAME="***" && \
export S3_CONFIG__PATH_PREFIX="***" && \
go run main.go

# example 2
export S3_CONFIG__REGION="us-west-2" && \
export S3_CONFIG__BUCKET_NAME="mys3bucketrsa2023" && \
export S3_CONFIG__PATH_PREFIX="/" && \
go run main.go
```