package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
)

type Response struct {
	Name    string
	Env     []string `json:",omitempty"`
	Request *RequestInfo
}

type RequestInfo struct {
	Method           string
	Header           http.Header
	URL              *url.URL
	Body             string
	ContentLength    int64
	TransferEncoding []string
	Host             string
	Form             url.Values
	PostForm         url.Values
	MultipartForm    *multipart.Form
	Trailer          http.Header
	RemoteAddr       string
	RequestURI       string
	UserAgent        string
	TLS              *tls.ConnectionState
}

func requestInfoFrom(req *http.Request) *RequestInfo {
	var err error
	var bodyString string
	defer req.Body.Close()
	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		log.Println("[ERROR] Error reading reqeust body: " + err.Error())
	} else {
		bodyString = string(bodyBytes)
	}

	log.Printf("[INFO] %q %q, Host: %q, Content Length: %d, %q, %q", req.Method, req.RequestURI, req.Host, req.ContentLength, req.UserAgent(), req.RemoteAddr)
	return &RequestInfo{
		Method:           req.Method,
		Header:           req.Header,
		URL:              req.URL,
		Body:             bodyString,
		ContentLength:    req.ContentLength,
		TransferEncoding: req.TransferEncoding,
		Host:             req.Host,
		Form:             req.Form,
		PostForm:         req.PostForm,
		MultipartForm:    req.MultipartForm,
		Trailer:          req.Trailer,
		RemoteAddr:       req.RemoteAddr,
		RequestURI:       req.RequestURI,
		UserAgent:        req.UserAgent(),
		TLS:              req.TLS,
	}
}

func reqInfo(w http.ResponseWriter, req *http.Request) {
	name, _ := os.LookupEnv("NAME")
	var env []string
	if readEnvs {
		env = os.Environ()
	}
	resp := Response{
		Name:    name,
		Env:     env,
		Request: requestInfoFrom(req),
	}

	var reqBytes []byte
	var err error
	if pretty {
		reqBytes, err = json.MarshalIndent(resp, "", "    ")
		if err != nil {
			log.Println("[ERROR] Error parsing JSON value, error: " + err.Error())
		}
	} else {
		reqBytes, err = json.Marshal(resp)
		if err != nil {
			log.Println("[ERROR] Error parsing JSON value, error: " + err.Error())
		}
	}

	_, _ = fmt.Fprintf(w, string(reqBytes))
}

func empty(w http.ResponseWriter, req *http.Request) {
	_, _ = fmt.Fprintf(w, "")
}

func main() {
	flag.BoolVar(&readEnvs, "read-envs", false, "Read environment variables")
	flag.BoolVar(&pretty, "pretty", false, "Prettify output JSON")
	flag.BoolVar(&https, "https", false, "HTTPS server")
	flag.StringVar(&cert, "cert", "", "Cert file for HTTPS server")
	flag.StringVar(&key, "key", "", "Key file for HTTPS server")
	flag.Parse()

	http.HandleFunc("/", reqInfo)
	http.HandleFunc("/empty-response", empty)
	log.Println("[INFO] Starting echo service...")
	log.Println("[INFO] Invoke resource '/' to echo response")
	log.Println("[INFO] Invoke '/empty-response' to return empty response")
	if https {
		log.Println("[INFO] Server listening at :8443")
		if err := http.ListenAndServeTLS(":8443", cert, key, nil); err != nil {
			panic(err)
		}
	} else {
		log.Println("[INFO] Server listening at :8080")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			panic(err)
		}
	}
}

var readEnvs bool
var pretty bool
var https bool
var cert string
var key string
