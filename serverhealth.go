package main

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
)

type WebsiteChecker func(string) bool

type Response struct {
	Metric   string `json: metric`
	Hostname string `json: hostname`
	Status   int `json: status`
}

var handlerMux sync.Mutex

func HealthCheck(url string) bool {
	response, err := http.Head(url)
	if err != nil {
		return false
	}
	return response.StatusCode == http.StatusOK
}

func DoHealthCheck(wc WebsiteChecker, hostname string) int {
	healthcheckMap := map[string]string{
        	"berryone": "http://192.168.0.110/admin/index.php",
	        "berrytwo": "http://192.168.0.120",
        	"berrythree": "http://192.168.0.130",
	        "berryfour": "http://192.168.0.140",
	}

	url, present := healthcheckMap[hostname]
	r := false
	if present {
		r = wc(url)
	} else {
		return -1
	}
	if !r {
		return 0
	} else {
		return 1
	}
}

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	hostname := strings.TrimPrefix(r.URL.Path, "/health/")
	if r.Method == http.MethodGet {
		handleGetPost(w, hostname)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleGetPost(w http.ResponseWriter, hostname string) {
	handlerMux.Lock()
	defer handlerMux.Unlock()
	health := DoHealthCheck(HealthCheck, hostname)
	if health < 0 {
		http.Error(w, "Host not found", http.StatusNotFound)
		return
	}

	stringResponse := fmt.Sprintf("healthcheck{hostname=%q} %v", hostname, health)
	w.Write([]byte(stringResponse))
}
