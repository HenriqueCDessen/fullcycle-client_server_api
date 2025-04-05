package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

const (
	apiURL     = "http://localhost:8081/cotacao"
	outputFile = "cotacao.txt"
	timeout    = 300 * time.Millisecond
)

type Response struct {
	Bid string `json:"bid"`
}

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	bid, err := getQuote(ctx)
	if err != nil {
		log.Fatalf("Erro ao obter cotação: %v", err)
	}

	if err = saveQuoteToFile(bid); err != nil {
		log.Fatalf("Erro ao salvar cotação: %v", err)
	}

	fmt.Println("Cotação salva com sucesso!")
}

func getQuote(ctx context.Context) (string, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, apiURL, nil)
	if err != nil {
		return "", fmt.Errorf("criação da requisição falhou: %w", err)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("requisição HTTP falhou: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("status code inválido (%d): %s", resp.StatusCode, string(body))
	}

	var response Response
	if err = json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return "", fmt.Errorf("decodificação JSON falhou: %w", err)
	}

	if response.Bid == "" {
		return "", fmt.Errorf("campo bid vazio na resposta")
	}

	return response.Bid, nil
}

func saveQuoteToFile(bid string) error {
	content := fmt.Sprintf("Dólar: %s\n", bid)
	return os.WriteFile(outputFile, []byte(content), 0644)
}
