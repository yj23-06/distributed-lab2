package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/rpc"
	"os"
	"uk.ac.bris.cs/distributed2/secretstrings/stubs"
)

func main() {
	server := flag.String("server", "127.0.0.1:8030", "IP:port string to connect to as server")
	flag.Parse()

	fmt.Println("Connecting to Server:", *server)

	client, err := rpc.Dial("tcp", *server)
	if err != nil {
		fmt.Printf("RPC Dial error: %v\n", err)
		return
	}
	defer client.Close()

	// Stage 1 Test
	singleReq := stubs.Request{Message: "University of Bristol"}
	var singleRes stubs.Response
	err = client.Call(stubs.ReverseHandler, singleReq, &singleRes)
	if err != nil {
		fmt.Printf("Stage 1 RPC error: %v\n", err)
	} else {
		fmt.Printf("Stage 1 Test: %s -> %s\n\n", singleReq.Message, singleRes.Message)
	}

	// Stage 3: Read words from wordlist file and reverse using FastReverse (Stage 2)
	wordlistPath := "../wordlist"
	file, err := os.Open(wordlistPath)
	if err != nil {
		// Fallback if client is run from root directory
		file, err = os.Open("wordlist")
		if err != nil {
			fmt.Printf("Error opening wordlist file: %v\n", err)
			return
		}
	}
	defer file.Close()

	fmt.Println("Processing wordlist via Premium FastReverse RPC...")
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		word := scanner.Text()
		if word == "" {
			continue
		}

		req := stubs.Request{Message: word}
		var res stubs.Response

		// Call Stage 2 Premium Endpoint
		err := client.Call(stubs.PremiumReverseHandler, req, &res)
		if err != nil {
			fmt.Printf("Error reversing '%s': %v\n", word, err)
			continue
		}

		fmt.Printf("%-20s -> %s\n", word, res.Message)
	}
}