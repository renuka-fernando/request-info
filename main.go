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
	"strconv"
	"time"
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

var serviceName string
var env []string

func handleRequest(w http.ResponseWriter, req *http.Request) *string {
	var err error
	var bodyString string
	defer req.Body.Close()
	bodyBytes, err := io.ReadAll(req.Body)
	if err != nil {
		log.Println("[ERROR] Error reading request body: " + err.Error())
	} else {
		bodyString = string(bodyBytes)
	}
	os.WriteFile("last_response", bodyBytes, 0644)
	log.Printf("[INFO] %q %q, Host: %q, Content Length: %d, %q, %q", req.Method, req.RequestURI, req.Host, req.ContentLength, req.UserAgent(), req.RemoteAddr)
	if logHeaders {
		log.Printf("[INFO] Headers: %v", req.Header)
	}
	if logBody {
		log.Printf("[INFO] Body: %v", bodyString)
	}

	respTime := responseTime
	respTimeStr := req.Header.Get("Set-Response-Time-Ms")
	if respTimeStr != "" {
		respTime, err = strconv.Atoi(respTimeStr)
		if err != nil {
			log.Println("[ERROR] Error parsing Set-Response-Time-Ms value, error: " + err.Error())
		}
	}

	code := statusCode
	codeStr := req.Header.Get("Set-Response-Status-Code")
	if codeStr != "" {
		code, err = strconv.Atoi(codeStr)
		if err != nil {
			log.Println("[ERROR] Error parsing Set-Response-Status-Code value, error: " + err.Error())
		}
	}

	time.Sleep(time.Duration(respTime) * time.Millisecond)
	w.WriteHeader(code)
	return &bodyString
}

func requestInfoFrom(w http.ResponseWriter, req *http.Request) *RequestInfo {
	bodyString := handleRequest(w, req)

	return &RequestInfo{
		Method:           req.Method,
		Header:           req.Header,
		URL:              req.URL,
		Body:             *bodyString,
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
	resp := Response{
		Name:    serviceName,
		Env:     env,
		Request: requestInfoFrom(w, req),
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

	_, _ = fmt.Fprintln(w, string(reqBytes))
}

func empty(w http.ResponseWriter, req *http.Request) {
	_ = handleRequest(w, req)
	_, _ = fmt.Fprintln(w, "")
}

func echo(w http.ResponseWriter, req *http.Request) {
	bodyString := handleRequest(w, req)
	_, _ = fmt.Fprintln(w, *bodyString)
}

func main() {
	flag.BoolVar(&readEnvs, "read-envs", false, "Read environment variables")
	flag.BoolVar(&pretty, "pretty", false, "Prettify output JSON")
	flag.BoolVar(&logHeaders, "logH", false, "Log Headers")
	flag.BoolVar(&logBody, "logB", false, "Log Headers")
	flag.BoolVar(&https, "https", false, "HTTPS server")
	flag.StringVar(&addr, "addr", "", "Address to bind the server")
	flag.StringVar(&cert, "cert", "", "Cert file for HTTPS server")
	flag.StringVar(&key, "key", "", "Key file for HTTPS server")
	flag.IntVar(&responseTime, "time", 0, "Time to wait (ms) before responding to request")
	flag.IntVar(&statusCode, "status", 200, "HTTP status code to respond")
	flag.Parse()

	serviceName, _ = os.LookupEnv("NAME")
	if readEnvs {
		env = os.Environ()
	}

	http.HandleFunc("/", reqInfo)
	http.HandleFunc("/empty", empty)
	http.HandleFunc("/echo", echo)
	log.Println("[INFO] Starting echo service...")
	log.Println("[INFO] Invoke resource '/' to get request info in response")
	log.Println("[INFO] Invoke resource '/echo' to echo response")
	log.Println("[INFO] Invoke '/empty' to return empty response")

	if addr == "" {
		if https {
			addr = ":8443"
		} else {
			addr = ":8080"
		}
	}
	log.Println("[INFO] Server listening at " + addr)

	if https {
		if err := http.ListenAndServeTLS(addr, cert, key, nil); err != nil {
			panic(err)
		}
	} else {
		if err := http.ListenAndServe(addr, nil); err != nil {
			panic(err)
		}
	}
}

var readEnvs bool
var pretty bool
var logHeaders bool
var logBody bool
var https bool
var addr string
var cert string
var key string
var responseTime int
var statusCode int
