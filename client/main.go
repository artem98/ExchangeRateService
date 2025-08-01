package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

func main() {
	addr := flag.String("addr", "http://localhost:8080", "Server address with port (e.g., http://localhost:8080)")
	flag.Parse()

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Client started. Now you can send commands. Send \"help\" to get full list of commands.")
	printHelp()
MAIN_LOOP:
	for {
		fmt.Print("> ")
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Error reading input:", err)
			continue
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, " ", 2)
		cmd := strings.ToLower(strings.TrimSpace(parts[0]))

		switch cmd {
		case "help":
			printHelp()
		case "post":
			if len(parts) < 2 {
				fmt.Println("Missing JSON body for POST.")
				continue
			}
			handlePost(*addr, strings.TrimSpace(parts[1]))
		case "get":
			if len(parts) < 2 {
				fmt.Println("Missing parameter for GET.")
				continue
			}
			handleGet(*addr, strings.TrimSpace(parts[1]))
		case "quit":
			break MAIN_LOOP
		default:
			fmt.Println("Unknown command:", cmd)
		}
	}
}

func handlePost(addr, jsonBody string) {
	url := fmt.Sprintf("%s/rates/update_requests", addr)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer([]byte(jsonBody)))
	if err != nil {
		fmt.Println("Failed to create POST request:", err)
		return
	}
	req.Header.Set("Content-Type", "application/json")

	sendAndPrint(req)
}

func handleGetByID(addr string, id uint64) {
	url := fmt.Sprintf("%s/rates/update_requests/%d", addr, id)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		fmt.Println("Failed to create GET by ID request:", err)
		return
	}
	sendAndPrint(req)
}

func handleGetByCurrencyPair(addr, pair string) {
	baseUrl := fmt.Sprintf("%s/rates", addr)
	params := url.Values{}
	params.Add("currency_pair", pair)

	fullUrl := fmt.Sprintf("%s?%s", baseUrl, params.Encode())
	req, err := http.NewRequest("GET", fullUrl, nil)
	if err != nil {
		fmt.Println("Failed to create GET by currency pair request:", err)
		return
	}
	sendAndPrint(req)
}

func handleGet(addr, param string) {
	if id, err := strconv.ParseUint(param, 10, 64); err == nil {
		handleGetByID(addr, id)
	} else {
		handleGetByCurrencyPair(addr, param)
	}
}

func sendAndPrint(req *http.Request) {
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Println("Request error:", err)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	fmt.Println("Response status:", resp.Status)
	fmt.Println("Response body:", string(body))
}

func printHelp() {
	fmt.Println("Available commands:")
	fmt.Println("  help                        - Show this help message.")
	fmt.Println("  quit                        - Quit.")
	fmt.Println("  get  <request_id>           - Get exchange rate by update request Id (uint64). Example: 'get 1234567'")
	fmt.Println("  post <json>                 - Post update request for currency pair. Example: 'post {\"currency_pair\":\"EUR/MXN\"}'")
	fmt.Println("  get  <currency_pair>        - Get exchange rate by currency pair (string). Example: 'get EUR/MXN'")
}
