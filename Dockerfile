FROM golang:1.20.1 AS builder
# Define build env
ENV GOOS linux
ENV CGO_ENABLED 0
# Add a work directory
WORKDIR /app
CMD pwd
# Cache and install dependencies
COPY ["go.mod", "./"]
COPY ["go.sum", "./"]
RUN go mod download
# Copy app files
COPY ["pkg/", "./pkg/"]
WORKDIR /app/cmd/scraper
COPY ["cmd/scraper/*.go", "./"]
# Build app
RUN go build -o scraper

FROM alpine:3.17.2 as runner
COPY ["cmd/scraper/config.yaml", "./"]
COPY --from=builder /app/cmd/scraper/scraper .
EXPOSE 8080
CMD ./scraper