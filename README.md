# URL_Shortener_Go
A service like tinyURL in Golang
## Steps to run
 * go run main.go
 * using curl for POST request. Replace long_url with the long url.
 
 
    ```curl -d '{"orig_url":"long_url"}' -H "Content-Type: application/json" -X POST http://localhost:8047/CREATE```
 * using curl for GET request. Replace short_url with the short url.
 
 
    ```curl http://localhost:8047/url/short_url```
