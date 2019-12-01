package main

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/big"
	"net/http"
	"strings"

	"github.com/philippgille/gokv/redis"
)
//struct for long url for accepting JSON
type ORG_URL_Struct struct {
	OrigURL string `json:"orig_url"`
}
//struct for short url for accepting JSON
type Short_URL_Struct struct {
	ShortURL string `json:"short_url"`
}

func get_short_url(n uint64) string {
	//map n to a 62 bit number
	//this will be the short URL
	alphabet := string("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	s := ""
	for n > 0 {
		s = s + alphabet[n%62:(n%62)+1]
		n = n / 62
	}
	return s
}
func createHandler(w http.ResponseWriter, r *http.Request) {
	//handling the post request
	org_url := ORG_URL_Struct{}
	jsn, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal("Error reading the body", err)
	}
	//converting json to the struct
	err = json.Unmarshal(jsn, &org_url)
	if err != nil {
		log.Fatal("Decoding error: ", err)
	}
	log.Printf("Received: %v\n", org_url)
	//calculating md5 hash value of the long url to get the id value
	//id_value will be passed to the get_short_url to get the 62 bit alphanumeric short url
	bi := big.NewInt(0)
	h := md5.New()
	h.Write([]byte(org_url.OrigURL))
	hexstr := hex.EncodeToString(h.Sum(nil))
	bi.SetString(hexstr, 16)
	id_value := bi.Uint64()
	s := get_short_url(id_value)
	short_url := Short_URL_Struct{
		ShortURL: s,
	}
	//creating the json with short url
	shortJson, err := json.Marshal(short_url)
	if err != nil {
		fmt.Fprintf(w, "Error: %s", err)
	}
	//store key value in redis
	options := redis.DefaultOptions // Address: "localhost:6379", Password: "", DB: 0
	// Create client to interact with redis server
	client, err := redis.NewClient(options)
	if err != nil {
		panic(err)
	}
	defer client.Close()
	//setting key s to value org_url.OrigURL, storing key value pair
	err_ := client.Set(s, org_url.OrigURL)
	if err_ != nil {
		panic(err)
	}
	retrievedVal := new(string)
	_, err__ := client.Get(s, retrievedVal)
	if err__ != nil {
		panic(err)
	}
	w.Header().Set("Content-Type", "application/json")
	//responding with json containing short-url
	w.Write(shortJson)

}
func redirectHandler(w http.ResponseWriter, r *http.Request) {
	message := r.URL.Path
	//extracting the short url
	message = strings.TrimPrefix(message, "/url/")
	fmt.Println("The short URL request is " + message)
	options := redis.DefaultOptions // Address: "localhost:6379", Password: "", DB: 0
	// Create client to redis server
	client, err := redis.NewClient(options)
	if err != nil {
		panic(err)
	}
	defer client.Close()
	retrievedVal := new(string)
	//get the value of the key message, basically a map from short url to long url
	ok, err__ := client.Get(message, retrievedVal)
	if err__ != nil {
		panic(err)
	}
	bb := *(retrievedVal)
	if ok == false {
		text := "URL is invalid"
		w.Write([]byte(text))
	}
	if ok {
		text := "Long URL is " + bb
		fmt.Println("The long URL is " + bb)
		w.Write([]byte(text))
	}
}
func server() {
	//handler for creating short url from long url and storing
	http.HandleFunc("/CREATE", createHandler)
	//handler for looking up if the short url is valid and returning long url if the short url is valid
	http.HandleFunc("/url/", redirectHandler)
	http.ListenAndServe(":8047", nil)
}

func main() {
	server()
}
