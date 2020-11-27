package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
)

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
	reqBytes, err := json.Marshal(requestInfoFrom(req))
	if err != nil {
		log.Println("Error parsing JSON value, error: " + err.Error())
	}
	_, _ = fmt.Fprintf(w, string(reqBytes))
}

func main() {
	http.HandleFunc("/", hello)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		panic(err)
	}
}
