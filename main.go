package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const KEY_FILENAME = ".ttkey"

const TOKEN_API = "https://api.cognitive.microsoft.com/sts/v1.0/issueToken"
const TRANSLATE_API = "https://api.microsofttranslator.com/v2/ajax.svc/Translate"

var ENREGEX = regexp.MustCompile("^[a-zA-Z0-9]")
var APPKEY []byte

type Token struct {
	Value   []byte `json:"token"`
	Expired int64  `json:"expired"`
}

func init() {
	keyFile := filepath.Join(os.Getenv("HOME"), KEY_FILENAME)

	if _, err := os.Stat(keyFile); err != nil {
		fmt.Printf("API key file ~/%s is not exists.\n", KEY_FILENAME)
		fmt.Println("Please generate follwing command:")
		fmt.Printf("echo [your-api-key] > ~/%s\n", KEY_FILENAME)
		os.Exit(1)
	}

	APPKEY, _ = ioutil.ReadFile(keyFile)
}

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Please pass argument translate word that you want.")
		os.Exit(1)
	}
	text := strings.Join(os.Args[1:], " ")
	from := "ja"
	to := "en"
	if ENREGEX.MatchString(text) {
		from = "en"
		to = "ja"
	}
	token := getToken(APPKEY)
	translated := translate(token, text, from, to)

	fmt.Println(strings.Replace(translated, "\"", "", -1))
}

func getToken(key []byte) *Token {
	var token *Token
	token = getTokenFromCache()
	if token != nil {
		return token
	}

	fmt.Println("Getting token...")
	url := fmt.Sprintf("%s?Subscription-Key=%s", TOKEN_API, strings.TrimSpace(string(key)))
	resp, err := http.Post(url, "", nil)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	if buf, err := ioutil.ReadAll(resp.Body); err != nil {
		panic(err)
	} else {
		t := &Token{
			Value:   buf,
			Expired: time.Now().Add(10 * time.Minute).Unix(),
		}
		if w, err := json.Marshal(t); err == nil {
			if err := ioutil.WriteFile("/tmp/tttoken", w, 0755); err != nil {
				fmt.Println("Parsed.")
			}
		} else {
			fmt.Println(err)
		}
		return t
	}
}

func getTokenFromCache() *Token {
	if _, err := os.Stat("/tmp/tttoken"); err != nil {
		return nil
	}

	buf, _ := ioutil.ReadFile("/tmp/tttoken")
	var t Token
	if err := json.Unmarshal(buf, &t); err != nil {
		return nil
	}

	if t.Expired < time.Now().Unix() {
		return nil
	}

	return &t
}

func translate(token *Token, text, from, to string) string {
	query := url.Values{}
	query.Add("text", text)
	query.Add("from", from)
	query.Add("to", to)
	query.Add("contentType", "text/plain")

	apiUrl := fmt.Sprintf("%s?%s", TRANSLATE_API, query.Encode())
	req, _ := http.NewRequest("GET", apiUrl, nil)
	req.Header.Add("Authorization", "Bearer "+string(token.Value))

	client := http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	ret, _ := ioutil.ReadAll(resp.Body)
	return string(ret)
}
