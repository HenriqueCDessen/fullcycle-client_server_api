package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

type CurrencyResponse struct {
	USDBRL struct {
		Bid string `json:"bid"`
	} `json:"USDBRL"`
}

type Response struct {
	Bid string `json:"bid"`
}

func main() {
	db, err := sql.Open("sqlite3", "cotacoes.db")
	if err != nil {
		log.Fatalf("Erro ao abrir o banco de dados: %v", err)
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS cotacoes (id INTEGER PRIMARY KEY, bid TEXT, timestamp DATETIME DEFAULT CURRENT_TIMESTAMP)`)
	if err != nil {
		log.Fatalf("Erro ao criar tabela: %v", err)
	}

	http.HandleFunc("/cotacao", func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		apiCtx, apiCancel := context.WithTimeout(ctx, 200*time.Millisecond)
		defer apiCancel()

		currency, err := fetchDollarQuote(apiCtx)
		if err != nil {
			http.Error(w, "Erro ao buscar cotação", http.StatusInternalServerError)
			log.Println("Erro na consulta à API:", err)
			return
		}

		dbCtx, dbCancel := context.WithTimeout(ctx, 10*time.Millisecond)
		defer dbCancel()

		err = saveQuote(dbCtx, db, currency.USDBRL.Bid)
		if err != nil {
			http.Error(w, "Erro ao salvar cotação", http.StatusInternalServerError)
			log.Println("Erro ao salvar no banco:", err)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(Response{Bid: currency.USDBRL.Bid})
	})

	log.Println("Servidor rodando na porta 8080...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func fetchDollarQuote(ctx context.Context) (*CurrencyResponse, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://economia.awesomeapi.com.br/json/last/USD-BRL", nil)
	if err != nil {
		return nil, err
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var currency CurrencyResponse
	err = json.NewDecoder(resp.Body).Decode(&currency)
	if err != nil {
		return nil, err
	}

	return &currency, nil
}

func saveQuote(ctx context.Context, db *sql.DB, bid string) error {
	query := `INSERT INTO cotacoes (bid) VALUES (?)`

	done := make(chan error)
	go func() {
		_, err := db.Exec(query, bid)
		done <- err
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		return err
	}
}
