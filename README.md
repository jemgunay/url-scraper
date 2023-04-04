# URL Scraper

## Usage

### Starting

```shell
cd cmd/scraper
go run main.go --config-path="config.yaml"
```

### Fetch URLs

```shell
curl -i -XGET http://localhost:8080/v1/api/urls
HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
Date: Tue, 04 Apr 2023 21:36:32 GMT
Content-Length: 2

[]
```

### Fetch URLs

```shell
curl -i -XPOST http://localhost:8080/v1/api/urls -d '{"url": "https://example.com"}'
HTTP/1.1 200 OK
Content-Type: application/json; charset=utf-8
Date: Tue, 04 Apr 2023 21:36:32 GMT
Content-Length: 2

[]
```