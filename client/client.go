package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Response struct {
	Bid float64 `json:"bid"`
}

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 300*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, "GET", "http://localhost:8080", nil)
	if err != nil {
		panic(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	if resp.StatusCode == http.StatusOK {
		defer resp.Body.Close()
		out, err := os.Create("cotacao.txt")
		if err != nil {
			panic(err)
		}
		defer out.Close()
		var data Response
		err = json.NewDecoder(resp.Body).Decode(&data)
		if err != nil {
			panic(err)
		}
		_, err = out.WriteString("Dólar: " + strconv.FormatFloat(data.Bid, 'f', -1, 64))
		if err != nil {
			panic(err)
		}
		io.Copy(os.Stdout, resp.Body)
	} else {
		log.Println("error on request")
	}
}
