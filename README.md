# URL Scraper

A toy service for storing, counting and periodically benchmarking URL request/response times. 

## Usage

### Running

#### Manual

```shell
cd cmd/scraper
go run main.go
go run main.go --config-path="config.yaml"
```

#### Docker

```shell
docker build -t jemgunay/url-scraper . --target runner
docker run -p 8080:8080 jemgunay/url-scraper
```

### Store URL

```shell
curl -i -XPOST 'http://localhost:8080/api/v1/urls' -d '{"url": "https://example.com"}'
HTTP/1.1 202 Accepted
Date: Wed, 05 Apr 2023 17:02:38 GMT
Content-Length: 0
```

You can also execute `./scripts/hydrate.sh` to hydrate the store with initial URLs.

### Fetch URLs

Returns 50 stored URLs. By default, returns URLs sorted by most recently submitted. 

```shell
curl -i -XGET 'http://localhost:8080/api/v1/urls'
HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
Date: Wed, 05 Apr 2023 17:20:35 GMT
Transfer-Encoding: chunked
[
  {"key":"https://httpbin.org/get?val=49","count":11,"last_upserted":"2023-04-05T17:20:25.426827Z"}
  {"key":"https://httpbin.org/get?val=43","count":9,"last_upserted":"2023-04-05T17:20:25.310556Z"},
  ...
]
```

```shell
# sortBy (age/count) & sortOrder (asc/desc) query params
curl -i -XGET 'http://localhost:8080/api/v1/urls?sortBy=age&sortOrder=asc'
curl -i -XGET 'http://localhost:8080/api/v1/urls?sortBy=count&sortOrder=desc'
```

### Example of 60s Scheduled URL Benchmarking

```json
{"level":"info","ts":"2023-04-05T19:50:50.125+0100","caller":"ingest/ingest.go:123","msg":"successfully refreshed URL benchmarks","summary": {
  "scrape_durations":[{"url":"https://httpbin.org/get?val=13","duration":"89.677ms","status":"success"},{"url":"https://httpbin.org/get?val=12","duration":"92.497ms","status":"success"},{"url":"https://httpbin.org/get?val=11","duration":"94.125ms","status":"success"},{"url":"https://httpbin.org/get?val=14","duration":"91.05ms","status":"success"},{"url":"https://httpbin.org/get?val=15","duration":"207.785ms","status":"success"},{"url":"https://httpbin.org/get?val=16","duration":"503.799ms","status":"success"},{"url":"https://httpbin.org/get?val=18","duration":"549.729ms","status":"success"},{"url":"https://httpbin.org/get?val=19","duration":"94.17ms","status":"success"},{"url":"https://httpbin.org/get?val=17","duration":"536.713ms","status":"success"},{"url":"https://httpbin.org/get?val=20","duration":"627.982ms","status":"success"
  "success_count":10,
  "failure_count":0}
}
```

### Run Tests

```shell
go test -race ./...
```

## TODO

* Write proper integration tests with `gomock` and improve unit test coverage.
* Properly structured hexagonal architecture with adapters, i.e. instead of packages accessing types in `ports.go`.
* Make more things configurable in yaml, e.g. daemon refresh frequency, queue capacities, etc.
* CI for build & test validation.
* Store benchmarks.