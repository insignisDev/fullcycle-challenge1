package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type CurrencySchema struct {
	gorm.Model
	Code       string
	Codein     string
	Name       string
	High       float64
	Low        float64
	VarBid     float64
	PctChange  float64
	Bid        float64
	Ask        float64
	Timestamp  string
	CreateDate string
}

type Currency struct {
	Code       string `json:"code"`
	Codein     string `json:"codein"`
	Name       string `json:"name"`
	High       string `json:"high"`
	Low        string `json:"low"`
	VarBid     string `json:"varBid"`
	PctChange  string `json:"pctChange"`
	Bid        string `json:"bid"`
	Ask        string `json:"ask"`
	Timestamp  string `json:"timestamp"`
	CreateDate string `json:"create_date"`
}

func dbConnect() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open("./test.db"), &gorm.Config{})
	if err != nil {
		return nil, err
	}
	db.AutoMigrate(&CurrencySchema{})
	return db, nil
}

func main() {
	db, err := dbConnect()
	if err != nil {
		panic(err)
	}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		select {
		case <-ctx.Done():
			if ctx.Err() == context.DeadlineExceeded {
				log.Println("Timeout fetching cotation from external API")
			} else {
				log.Println("Error fetching cotation:", ctx.Err())
			}
		default:
			handler(w, r, db)
		}
	})

	http.ListenAndServe(":8080", nil)
}

func handler(w http.ResponseWriter, r *http.Request, db *gorm.DB) {
	cotation, err := getCotation(db)
	if err != nil {
		log.Fatalln("timeout client request")
		fmt.Fprintf(w, "{timeout}", "timeout")
		// http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	fmt.Fprintf(w, "{\"Bid\": %.2f}", parseFloat(cotation.Bid))
}

func parseFloat(val string) float64 {
	parseFloat, err := strconv.ParseFloat(val, 64)
	if err != nil {
		panic(err)
	}
	return parseFloat
}

func saveCotation(db *gorm.DB, cotation Currency) {
	high, err := strconv.ParseFloat(cotation.High, 64)
	if err != nil {
		panic(err)
	}
	currency := CurrencySchema{
		Code:       cotation.Code,
		Codein:     cotation.Codein,
		Name:       cotation.Name,
		High:       high,
		Low:        parseFloat(cotation.Low),
		VarBid:     parseFloat(cotation.VarBid),
		PctChange:  parseFloat(cotation.PctChange),
		Bid:        parseFloat(cotation.Bid),
		Ask:        parseFloat(cotation.Ask),
		Timestamp:  cotation.Timestamp,
		CreateDate: cotation.CreateDate,
	}
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 0*time.Millisecond)
	defer cancel()
	db.WithContext(ctx).Save(&currency)

}

func getCotation(db *gorm.DB) (Currency, error) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
	defer cancel()
	select {
	case <-ctx.Done():
		log.Println("timeout client request")
		return Currency{}, ctx.Err()
	default:
		req, err := http.NewRequestWithContext(ctx, "GET", "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
		if err != nil {
			return Currency{}, err
		}
		res, err := http.DefaultClient.Do(req)
		if err != nil {
			return Currency{}, err
		}
		defer res.Body.Close()
		var result map[string]Currency
		err = json.NewDecoder(res.Body).Decode(&result)
		if err != nil {
			return Currency{}, err
		}
		cotation := result["USDBRL"]
		saveCotation(db, cotation)
		return cotation, nil
	}
}
