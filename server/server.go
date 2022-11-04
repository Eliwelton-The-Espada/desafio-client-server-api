package main

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type (
	Output struct {
		Usdbrl Usdbrl `json:"USDBRL"`
	}
	Usdbrl struct {
		ID         int    `gorm:"primaryKey"`
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
)

type OutputClient struct {
	Bid string `json:"bid"`
}

func main() {
	http.HandleFunc("/cotacao", getCotacaoHandler)
	log.Println("Listening on port 8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func getCotacaoHandler(w http.ResponseWriter, r *http.Request) {
	out, err := getCotacao()
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	outb, err := json.Marshal(out)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write(outb)
}

func getCotacao() (outCli OutputClient, err error) {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 200*time.Millisecond)
	defer cancel()

	url := "https://economia.awesomeapi.com.br/json/last/USD-BRL"

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return
	}

	log.Println("Payload received:", string(body))

	var usdbrl Output
	err = json.Unmarshal(body, &usdbrl)
	if err != nil {
		return
	}

	err = saveOnDatabase(usdbrl.Usdbrl)
	if err != nil {
		return
	}

	outCli = OutputClient{
		Bid: usdbrl.Usdbrl.Bid,
	}
	return
}

func saveOnDatabase(usdbrl Usdbrl) error {
	ctx := context.Background()
	ctx, cancel := context.WithTimeout(ctx, 10*time.Millisecond)
	defer cancel()

	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		return err
	}

	err = db.AutoMigrate(&Usdbrl{})
	if err != nil {
		return err
	}

	db.WithContext(ctx).Create(&usdbrl)

	// Por algum motivo o ID (primary Key) tá sempre iniciando com 0 (Zero)
	// Devido a isso ele fica sem salvar o registro da primeira request
	// E a partir da segunda request é que ele começa a salvar porque o ID vai para 1
	// Não sei se é devido ao SQLite, pois quando eu estava mexendo com o MySQL ele sempre tava iniciando em 1
	// Deixei um log para uma melhor análise, ele sempre pega o registro do ID atual e printa na tela
	var out Usdbrl
	db.First(&out, usdbrl.ID)
	log.Println("Saved on database:", out)

	return nil
}
