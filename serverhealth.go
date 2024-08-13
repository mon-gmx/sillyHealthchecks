package main

import (
	//"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

type WebsiteChecker func(string) bool

type HostMetrics struct {
	url    string
	status bool
}

type Response struct {
	Metric   string `json: metric`
	Hostname string `json: hostname`
	Status   string `json: status`
}

var handlerMux sync.Mutex

func HealthCheck(url string) bool {
	response, err := http.Head(url)
	if err != nil {
		return false
	}
	return response.StatusCode == http.StatusOK
}

func CheckWebsites(wc WebsiteChecker, urls []string) map[string]bool {
	results := make(map[string]bool)
	resultChannel := make(chan HostMetrics)

	for _, url_ := range urls {
		go func(u string) {
			u = "http://" + u
			parsedURL, err := url.Parse(u)
			if err == nil {
				hostname := parsedURL.Hostname()
				log.Printf("hostname: %v, url: %v", hostname, u)
				resultChannel <- HostMetrics{url: hostname, status: wc(u)}
			}
		}(url_)

	}

	for i := 0; i < len(urls); i++ {
		r := <-resultChannel
		results[r.url] = r.status
	}
	return results
}

func ServeHTTP(w http.ResponseWriter, r *http.Request) {
	host := strings.TrimPrefix(r.URL.Path, "/health/")
	log.Println(r)
	if r.Method == http.MethodGet {
		handleGetPost(w, host)
	} else {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func handleGetPost(w http.ResponseWriter, host string) {
	handlerMux.Lock()
	defer handlerMux.Unlock()
	status := "bad"

	urls := []string{
		"192.168.0.110/admin/index.php",
		"192.168.0.120",
		"192.168.0.130",
		"192.168.0.140",
	}
	healthMap := CheckWebsites(HealthCheck, urls)
	health, ok := healthMap[host]
	if !ok {
		http.Error(w, "Host not found", http.StatusNotFound)
	}
	if health {
		status = "ok"
	}

	/*
		response := Response{
			Hostname: host,
			Status:   status,
			Metric:   "healtcheck",
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(response)
	*/

	// formatting as prometheus response
	stringResponse := fmt.Sprintf("healthcheck{host=%q} %v", host, status)
	w.Write([]byte(stringResponse))
}
