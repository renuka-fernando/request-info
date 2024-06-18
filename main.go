package main

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"sync"
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

const delayMsSeparator = "-"

var serviceName string
var env []string

var muResp = &sync.RWMutex{}
var isResponseDataSet bool = false
var responseData string = ""

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

	respTime := delayMs
	respTimeStr := req.URL.Query().Get("delayMs")
	if respTimeStr != "" {
		// handle random time if "-" is present
		if strings.Contains(respTimeStr, delayMsSeparator) {
			randTime := strings.Split(respTimeStr, delayMsSeparator)
			randTimeMin, err := strconv.Atoi(randTime[0])
			if err != nil {
				log.Println("[ERROR] Error parsing Set-Response-Time-Ms value, error: " + err.Error())
			}
			randTimeMax, err := strconv.Atoi(randTime[1])
			if err != nil {
				log.Println("[ERROR] Error parsing Set-Response-Time-Ms value, error: " + err.Error())
			}
			rand.Seed(time.Now().UnixNano())
			respTime = rand.Intn(randTimeMax-randTimeMin) + randTimeMin
		} else {
			respTime, err = strconv.Atoi(respTimeStr)
			if err != nil {
				log.Println("[ERROR] Error parsing Set-Response-Time-Ms value, error: " + err.Error())
			}
		}
	}

	code := statusCode
	codeStr := req.URL.Query().Get("statusCode")
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
	w.Header().Set("Content-Type", "application/json")

	resp := Response{
		Name:    serviceName,
		Env:     env,
		Request: requestInfoFrom(w, req),
	}

	var reqBytes []byte
	var err error
	doPretty := pretty
	prettyStr := req.URL.Query().Get("pretty")
	if prettyStr != "" {
		doPretty, err = strconv.ParseBool(prettyStr)
		if err != nil {
			log.Println("[ERROR] Error parsing pretty value, error: " + err.Error())
		}
	}
	if doPretty {
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

func reqInfoSetPayloadHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	reqPayload := handleRequest(w, req)

	if req.Method == "POST" && reqPayload != nil {
		log.Println("[INFO] Setting response data to: " + *reqPayload)

		muResp.Lock()
		isResponseDataSet = true
		responseData = *reqPayload
		muResp.Unlock()
		fmt.Fprintln(w, `{"message":"Response data is set successfully"}`)
	} else if req.Method == "DELETE" {
		muResp.Lock()
		isResponseDataSet = false
		muResp.Unlock()
		fmt.Fprintln(w, `{"message":"Response data is removed successfully"}`)
	}

}

func setResponseHandler(handler func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		muResp.RLock()
		defer muResp.RUnlock()
		if isResponseDataSet {
			_, _ = fmt.Fprintln(w, responseData)
		} else {
			handler(w, req) // call the original handler
		}
	}
}

func healthzHandler(w http.ResponseWriter, req *http.Request) {
	w.WriteHeader(http.StatusOK)
	_, _ = fmt.Fprint(w, "OK")
}

func empty(w http.ResponseWriter, req *http.Request) {
	_ = handleRequest(w, req)
	_, _ = fmt.Fprint(w, "")
}

func echo(w http.ResponseWriter, req *http.Request) {
	bodyString := handleRequest(w, req)
	_, _ = fmt.Fprintln(w, *bodyString)
}

func enableCORS(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, HEAD, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// If it's a preflight request, send the necessary headers and stop the request chain
		if r.Method == "OPTIONS" {
			return
		}

		// Call the actual handler
		handler.ServeHTTP(w, r)
	})
}

func main() {
	flag.BoolVar(&readEnvs, "read-envs", false, "Read environment variables")
	flag.BoolVar(&pretty, "pretty", false, "Prettify output JSON")
	flag.BoolVar(&logHeaders, "logH", false, "Log Headers")
	flag.BoolVar(&logBody, "logB", false, "Log Headers")
	flag.BoolVar(&https, "https", false, "HTTPS server") // Set to true for mTLS
	flag.BoolVar(&mtls, "mtls", false, "Enable mTLS")    // New flag to enable/disable mTLS
	flag.StringVar(&addr, "addr", "", "Address to bind the server")
	flag.StringVar(&cert, "cert", "server.crt", "Cert file for HTTPS server")
	flag.StringVar(&key, "key", "server.key", "Key file for HTTPS server")
	flag.StringVar(&clientCA, "ca", "ca.crt", "CA certificate file for client verification")
	flag.IntVar(&delayMs, "delayMs", 0, "Time to wait (ms) before responding to request")
	flag.IntVar(&statusCode, "status", 200, "HTTP status code to respond")
	flag.Parse()

	serviceName, _ = os.LookupEnv("NAME")
	if readEnvs {
		env = os.Environ()
	}

	// Create a new HTTP server
	server := http.NewServeMux()

	// Set up a handler for the desired route
	server.HandleFunc("/req-info/response", reqInfoSetPayloadHandler)
	server.HandleFunc("/empty", setResponseHandler(empty))
	server.HandleFunc("/echo", setResponseHandler(echo))
	server.HandleFunc("/healthz", healthzHandler)
	server.HandleFunc("/", setResponseHandler(reqInfo))

	// Enable CORS middleware
	corsHandler := enableCORS(server)

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
		tlsConfig := &tls.Config{
			MinVersion:   tls.VersionTLS12,
			Certificates: make([]tls.Certificate, 1),
		}
		tlsConfig.Certificates[0], _ = tls.LoadX509KeyPair(cert, key)

		// Load CA certificate for client verification
		caCert, _ := os.ReadFile(clientCA)
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		if mtls {
			tlsConfig.ClientCAs = caCertPool
			tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
		} else {
			tlsConfig.ClientCAs = nil
			tlsConfig.ClientAuth = tls.NoClientCert
		}

		serverTLS := &http.Server{
			Addr:      addr,
			Handler:   corsHandler,
			TLSConfig: tlsConfig,
		}

		// Start HTTPS server
		if err := serverTLS.ListenAndServeTLS("", ""); err != nil {
			log.Fatal(err)
		}
	} else {
		if err := http.ListenAndServe(addr, corsHandler); err != nil {
			log.Fatal(err)
		}
	}
}

var readEnvs bool
var pretty bool
var logHeaders bool
var logBody bool
var https bool
var mtls bool
var addr string
var cert string
var key string
var clientCA string
var delayMs int
var statusCode int
