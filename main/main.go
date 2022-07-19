package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"

	address "github.com/filecoin-project/go-address"
	"github.com/gorilla/mux"

	"github.com/filecoin-project/saturn-l2/resources"
)

type config struct {
	FilAddr string `json:"fil_wallet_address"`
}

func main() {
	var port int
	portStr := os.Getenv("PORT")
	if portStr == "" {
		port = 5500
	} else {
		var err error
		port, err = strconv.Atoi(portStr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Invalid PORT value '%s': %s\n", portStr, err.Error())
			os.Exit(1)
		}
	}

	filAddr := os.Getenv("FIL_WALLET_ADDRESS")
	if filAddr == "" {
		fmt.Fprintf(os.Stderr, "No FIL_WALLET_ADDRESS provided. Please set the environment variable.\n")
		os.Exit(2)
	}
	if _, err := address.NewFromString(filAddr); err != nil {
		fmt.Fprintf(os.Stderr, "Invalid FIL_WALLET_ADDRESS format: %s\n", err.Error())
		os.Exit(3)
	}
	cfg := config{FilAddr: filAddr}
	cfgJson, err := json.Marshal(cfg)
	if err != nil {
		panic(errors.New("failed to serialize config"))
	}

	m := mux.NewRouter()
	m.PathPrefix("/config").Handler(http.HandlerFunc(configHandler(cfgJson)))
	m.PathPrefix("/webui").Handler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		webuiHandler(cfg, w, r)
	}))

	srv := &http.Server{
		Handler: m,
	}

	nl, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot start the webserver: %s\n", err.Error())
		os.Exit(4)
	}

	port = nl.Addr().(*net.TCPAddr).Port
	fmt.Println("Server listening on", nl.Addr())
	fmt.Printf("WebUI: http://localhost:%d/webui\n", port)

	if err := srv.Serve(nl); err != http.ErrServerClosed {
		panic(err)
	}
}

func webuiHandler(cfg config, w http.ResponseWriter, r *http.Request) {
	rootDir := "webui"
	path := strings.TrimPrefix(r.URL.Path, "/")

	if path == rootDir {
		targetUrl := fmt.Sprintf("/%s/address/%s", rootDir, cfg.FilAddr)
		statusCode := 303 // See Other (a temporary redirect)
		http.Redirect(w, r, targetUrl, statusCode)
		return
	}

	_, err := resources.WebUI.Open(path)
	if path == rootDir || os.IsNotExist(err) {
		// file does not exist, serve index.html
		index, err := resources.WebUI.ReadFile(rootDir + "/index.html")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write(index)
		return
	} else if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// otherwise, use http.FileServer to serve the static dir
	http.FileServer(http.FS(resources.WebUI)).ServeHTTP(w, r)
}

func configHandler(conf []byte) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write(conf)
	}
}
