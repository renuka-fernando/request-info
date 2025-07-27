package main

import (
	"bytes"
	"context"
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
	"os/exec"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"
)

var healthy int32

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

	if !disableAccessLogs {
		log.Printf("[INFO] %q %q, Host: %q, Content Length: %d, %q, %q", req.Method, req.RequestURI, req.Host, req.ContentLength, req.UserAgent(), req.RemoteAddr)
	}
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

func setResponseBodyFromArgsHandler(w http.ResponseWriter, req *http.Request) {
	defer req.Body.Close()
	_, err := io.ReadAll(req.Body)
	if err != nil {
		log.Println("[ERROR] Error reading request body: " + err.Error())
	}
	fmt.Fprintln(w, responseBody)
}

func healthzHandler(w http.ResponseWriter, req *http.Request) {
	if atomic.LoadInt32(&healthy) == 1 {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK")
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
		fmt.Fprintf(w, "Service Unavailable")
	}
}

func empty(w http.ResponseWriter, req *http.Request) {
	_ = handleRequest(w, req)
	_, _ = fmt.Fprint(w, "")
}

func echo(w http.ResponseWriter, req *http.Request) {
	bodyString := handleRequest(w, req)
	_, _ = fmt.Fprintln(w, *bodyString)
}

func readFile(w http.ResponseWriter, req *http.Request) {
	_ = handleRequest(w, req)

	// Get the path parameter from the query string
	path := req.URL.Query().Get("path")
	if path == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = fmt.Fprintln(w, `{"error": "path parameter is required"}`)
		return
	}

	// Read the file
	content, err := os.ReadFile(path)
	if err != nil {
		log.Printf("[ERROR] Error reading file %s: %v", path, err)
		w.WriteHeader(http.StatusNotFound)
		_, _ = fmt.Fprintf(w, `{"error": "Failed to read file: %v"}`, err)
		return
	}

	// Return the file content
	_, _ = fmt.Fprint(w, string(content))
}

func listFiles(w http.ResponseWriter, req *http.Request) {
	_ = handleRequest(w, req)

	// Get the path parameter from the query string
	path := req.URL.Query().Get("path")
	if path == "" {
		path = "." // Default to current directory if no path provided
	}

	// Read the directory
	entries, err := os.ReadDir(path)
	if err != nil {
		log.Printf("[ERROR] Error reading directory %s: %v", path, err)
		w.WriteHeader(http.StatusNotFound)
		_, _ = fmt.Fprintf(w, `{"error": "Failed to read directory: %v"}`, err)
		return
	}

	// Build the response
	var files []string
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() {
			name += "/"
		}
		files = append(files, name)
	}

	// Return as JSON
	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"path":  path,
		"files": files,
	}

	jsonBytes, err := json.MarshalIndent(response, "", "    ")
	if err != nil {
		log.Printf("[ERROR] Error marshaling JSON: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprintf(w, `{"error": "Failed to marshal JSON: %v"}`, err)
		return
	}

	_, _ = fmt.Fprint(w, string(jsonBytes))
}

func executeCommand(w http.ResponseWriter, req *http.Request) {
	bodyString := handleRequest(w, req)

	// Only allow POST method
	if req.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		_, _ = fmt.Fprintln(w, `{"error": "Only POST method is allowed"}`)
		return
	}

	// Get the command from request body
	command := strings.TrimSpace(*bodyString)
	if command == "" {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = fmt.Fprintln(w, `{"error": "Command is required in request body"}`)
		return
	}

	log.Printf("[INFO] Executing command: %s", command)

	// Execute the command using shell
	cmd := exec.Command("sh", "-c", command)

	// Capture both stdout and stderr
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Execute the command
	err := cmd.Run()

	// Get the exit code
	exitCode := 0
	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			exitCode = 1
		}
	}

	// Prepare response
	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"command":   command,
		"stdout":    stdout.String(),
		"stderr":    stderr.String(),
		"exit_code": exitCode,
		"success":   exitCode == 0,
	}

	jsonBytes, jsonErr := json.MarshalIndent(response, "", "    ")
	if jsonErr != nil {
		log.Printf("[ERROR] Error marshaling JSON: %v", jsonErr)
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = fmt.Fprintf(w, `{"error": "Failed to marshal JSON: %v"}`, jsonErr)
		return
	}

	_, _ = fmt.Fprint(w, string(jsonBytes))
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
	flag.BoolVar(&readEnvs, "read-envs", true, "Read environment variables")
	flag.BoolVar(&pretty, "pretty", false, "Prettify output JSON")
	flag.BoolVar(&logHeaders, "logH", false, "Log Headers")
	flag.BoolVar(&logBody, "logB", false, "Log Body")
	flag.BoolVar(&https, "https", false, "HTTPS server") // Set to true for mTLS
	flag.BoolVar(&mtls, "mtls", false, "Enable mTLS")    // New flag to enable/disable mTLS
	flag.StringVar(&addr, "addr", "", "Address to bind the server")
	flag.StringVar(&cert, "cert", "server.crt", "Cert file for HTTPS server")
	flag.StringVar(&key, "key", "server.key", "Key file for HTTPS server")
	flag.StringVar(&clientCA, "ca", "ca.crt", "CA certificate file for client verification")
	flag.IntVar(&delayMs, "delayMs", 0, "Time to wait (ms) before responding to request")
	flag.IntVar(&statusCode, "statusCode", 200, "HTTP status code to respond")
	flag.BoolVar(&setResponseBody, "setResponseBody", false, "Response the body set via -responseBody flag")
	flag.StringVar(&responseBody, "responseBody", "", "The response body to return")
	flag.BoolVar(&disableAccessLogs, "disable-access-logs", false, "Disable access logs")
	flag.IntVar(&waitBeforeGracefulShutdownMs, "wait-before-graceful-shutdown-ms", 0, "Time to wait (ms) before graceful shutdown")
	flag.Parse()

	// Initially, the server is healthy
	atomic.StoreInt32(&healthy, 1)

	serviceName, _ = os.LookupEnv("NAME")
	if readEnvs {
		env = os.Environ()
	}

	// Channel to listen for OS signals
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	// Create a new HTTP server
	server := http.NewServeMux()

	server.HandleFunc("/healthz", healthzHandler)

	if setResponseBody {
		server.HandleFunc("/", setResponseBodyFromArgsHandler)
	} else {
		server.HandleFunc("/req-info/response", reqInfoSetPayloadHandler)
		server.HandleFunc("/empty", setResponseHandler(empty))
		server.HandleFunc("/echo", setResponseHandler(echo))
		server.HandleFunc("/file/read", setResponseHandler(readFile))
		server.HandleFunc("/file/list", setResponseHandler(listFiles))
		server.HandleFunc("/command", executeCommand)
		server.HandleFunc("/", setResponseHandler(reqInfo))
	}

	// Enable CORS middleware
	corsHandler := enableCORS(server)

	log.Println("[INFO] Starting server...")

	if setResponseBody {
		log.Println("[INFO] Responding the body set via -responseBody flag")
	} else {
		log.Println("[INFO] Invoke resource '/' to get request info in response")
		log.Println("[INFO] Invoke resource '/echo' to echo response")
		log.Println("[INFO] Invoke '/empty' to return empty response")
		log.Println("[INFO] Invoke '/file/read?path=<path>' to read a file from local directory")
		log.Println("[INFO] Invoke '/file/list?path=<path>' to list files in directory")
		log.Println("[INFO] POST to '/command' with command in body to execute shell commands")
	}

	if addr == "" {
		if https {
			addr = ":8443"
		} else {
			addr = ":8080"
		}
	}

	srv := &http.Server{
		Addr:         addr,
		Handler:      corsHandler,
		IdleTimeout:  180 * time.Second,
		ReadTimeout:  60 * time.Second,
		WriteTimeout: 60 * time.Second,
	}

	if https {
		log.Println("[INFO] Starting HTTPS server...")
		tlsConfig := &tls.Config{
			MinVersion:   tls.VersionTLS12,
			Certificates: make([]tls.Certificate, 1),
		}
		var err error
		tlsConfig.Certificates[0], err = tls.LoadX509KeyPair(cert, key)
		if err != nil {
			log.Fatalf("[ERROR] Error loading TLS certificate: %v", err)
		}

		if mtls {
			log.Println("[INFO] mTLS enabled")
			// Load CA certificate for client verification
			caCert, err := os.ReadFile(clientCA)
			if err != nil {
				log.Fatalf("Error loading CA certificate: %v", err)
			}

			caCertPool := x509.NewCertPool()
			caCertPool.AppendCertsFromPEM(caCert)

			tlsConfig.ClientCAs = caCertPool
			tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
		} else {
			tlsConfig.ClientCAs = nil
			tlsConfig.ClientAuth = tls.NoClientCert
		}

		srv.TLSConfig = tlsConfig
	}

	// Start HTTP server
	go func() {
		if https {
			if err := srv.ListenAndServeTLS("", ""); err != nil && err != http.ErrServerClosed {
				log.Fatalf("Error starting the HTTPS server: %v", err)
			}
		} else if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error starting the HTTP server: %v", err)
		}
	}()

	log.Println("[INFO] Server listening at " + addr)

	// Block until we receive a signal
	<-quit
	log.Println("Shutting down server...")

	// Set the server as unhealthy
	atomic.StoreInt32(&healthy, 0)

	// Create a context with a timeout for the shutdown process
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Wait before graceful shutdown
	if waitBeforeGracefulShutdownMs > 0 {
		log.Printf("Waiting %d ms before graceful shutdown...", waitBeforeGracefulShutdownMs)
		time.Sleep(time.Duration(waitBeforeGracefulShutdownMs) * time.Millisecond)
	}
	log.Println("Attempting graceful shutdown...")

	// Attempt graceful shutdown
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exiting")
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
var setResponseBody bool
var responseBody string
var disableAccessLogs bool
var waitBeforeGracefulShutdownMs int
