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

type ORG_URL_Struct struct {
	OrigURL string `json:"orig_url"`
}
type Short_URL_Struct struct {
	ShortURL string `json:"short_url"`
}

var m map[string]string
var ini_value int

func get_short_url(n uint64) string {
	// fmt.Println(n)
	alphabet := string("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	s := ""
	for n > 0 {
		s = s + alphabet[n%62:(n%62)+1]
		n = n / 62
	}
	return s
}
func createHandler(w http.ResponseWriter, r *http.Request) {
	org_url := ORG_URL_Struct{}
	jsn, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatal("Error reading the body", err)
	}
	err = json.Unmarshal(jsn, &org_url)
	if err != nil {
		log.Fatal("Decoding error: ", err)
	}
	log.Printf("Received: %v\n", org_url)

	//calculating md5 hash value of the long url to get the id value
	bi := big.NewInt(0)
	h := md5.New()
	h.Write([]byte(org_url.OrigURL))
	hexstr := hex.EncodeToString(h.Sum(nil))
	bi.SetString(hexstr, 16)
	// fmt.Println(bi.String())
	id_value := bi.Uint64()
	s := get_short_url(id_value)
	short_url := Short_URL_Struct{
		ShortURL: s,
	}
	shortJson, err := json.Marshal(short_url)

	if err != nil {
		fmt.Fprintf(w, "Error: %s", err)
	}
	// fmt.Println("here is the original url")
	// fmt.Println(org_url.OrigURL)

	m[s] = org_url.OrigURL
	//store key value in redis
	options := redis.DefaultOptions // Address: "localhost:6379", Password: "", DB: 0
	// Create client
	client, err := redis.NewClient(options)
	if err != nil {
		panic(err)
	}
	defer client.Close()
	err_ := client.Set(s, org_url.OrigURL)
	if err_ != nil {
		panic(err)
	}
	retrievedVal := new(string)
	_, err__ := client.Get(s, retrievedVal)
	if err__ != nil {
		panic(err)
	}

	// bb := *(retrievedVal)
	// fmt.Println(bb)
	w.Header().Set("Content-Type", "application/json")
	w.Write(shortJson)

}
func redirectHandler(w http.ResponseWriter, r *http.Request) {
	message := r.URL.Path
	message = strings.TrimPrefix(message, "/url/")
	fmt.Println("The short URL request is " + message)
	_, ok := m[message]
	options := redis.DefaultOptions // Address: "localhost:6379", Password: "", DB: 0
	// Create client to redis server
	client, err := redis.NewClient(options)
	if err != nil {
		panic(err)
	}
	defer client.Close()
	retrievedVal := new(string)
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
	http.HandleFunc("/CREATE", createHandler)
	http.HandleFunc("/url/", redirectHandler)
	http.ListenAndServe(":8047", nil)
}

func main() {
	m = make(map[string]string)
	ini_value = 10000
	server()
}
