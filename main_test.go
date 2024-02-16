package main

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandler(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(Handler))
	defer server.Close()

	resp, err := http.Get(server.URL + "/clientes/1/transacoes")

	if err != nil {
		t.Fatalf("error making request to server. Err: %v", err)
	}

	defer resp.Body.Close()

	// Assertions
	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected status OK; got %v", resp.Status)
	}
}
