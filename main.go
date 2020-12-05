package main

import (
	"crypto/tls"
	"encoding/json"
	"flag"
	"fmt"
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
	ContentLength    int64
	TransferEncoding []string
	Host             string
	Form             url.Values
	PostForm         url.Values
	MultipartForm    *multipart.Form
	Trailer          http.Header
	RemoteAddr       string
	RequestURI       string
	TLS              *tls.ConnectionState
}

func requestInfoFrom(req *http.Request) *RequestInfo {
	return &RequestInfo{
		Method:           req.Method,
		Header:           req.Header,
		URL:              req.URL,
		ContentLength:    req.ContentLength,
		TransferEncoding: req.TransferEncoding,
		Host:             req.Host,
		Form:             req.Form,
		PostForm:         req.PostForm,
		MultipartForm:    req.MultipartForm,
		Trailer:          req.Trailer,
		RemoteAddr:       req.RemoteAddr,
		RequestURI:       req.RequestURI,
		TLS:              req.TLS,
	}
}

func hello(w http.ResponseWriter, req *http.Request) {
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
			log.Println("Error parsing JSON value, error: " + err.Error())
		}
	} else {
		reqBytes, err = json.Marshal(resp)
		if err != nil {
			log.Println("Error parsing JSON value, error: " + err.Error())
		}
	}

	_, _ = fmt.Fprintf(w, string(reqBytes))
}

func main() {
	flag.BoolVar(&readEnvs, "read-envs", false, "Read environment variables")
	flag.BoolVar(&pretty, "pretty", false, "Prettify output JSON")
	flag.Parse()

	http.HandleFunc("/", hello)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}

var readEnvs bool
var pretty bool
