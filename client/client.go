package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type Response struct {
	Bid string `json:"bid"`
}

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 300*time.Millisecond)
	defer cancel()

	url := "http://localhost:8080/cotacao"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		panic(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	fmt.Println("Request header received:", resp.Header)
	fmt.Println("Payload received:", string(body))

	if string(body) != "" {
		var data Response
		err = json.Unmarshal(body, &data)
		if err != nil {
			panic(err)
		}

		f, err := os.Create("cotacao.txt")
		if err != nil {
			panic(err)
		}
		defer f.Close()

		formatString := fmt.Sprintf("Dólar: %s", data.Bid)

		size, err := f.WriteString(formatString)
		if err != nil {
			panic(err)
		}

		fmt.Printf("File created successfully! Size: %d bytes\n", size)
	}
}
