package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"

	_ "github.com/joho/godotenv/autoload"
	_ "github.com/mattn/go-sqlite3"
)

var (
	db       *sql.DB
	port     = os.Getenv("PORT")
	dburl    = os.Getenv("DB_URL")
	validate = validator.New()
)

type Cliente struct {
	Limite int64 `json:"limite"`
	Saldo  int64 `json:"saldo"`
}

type Transacao struct {
	Valor       int64  `json:"valor" validate:"gte=0"`
	Tipo        string `json:"tipo" validate:"oneof=c d"`
	Descricao   string `json:"descricao" validate:"min=1,max=10"`
	RealizadaEm string `json:"realizada_em"`
}

type Saldo struct {
	Total       int64  `json:"total"`
	Limite      int64  `json:"limite"`
	DataExtrato string `json:"data_extrato"`
}

type Extrato struct {
	Saldo             Saldo     `json:"saldo"`
	UltimasTransacoes []Transacao `json:"ultimas_transacoes"`
}

func Handler(w http.ResponseWriter, r *http.Request) {
	var transacao Transacao

	clienteId, _ := strconv.Atoi(r.URL.Path[:strings.IndexByte(r.URL.Path, '/')])
	endpoint := r.URL.Path[strings.IndexByte(r.URL.Path, '/')+1:]

	switch {
	case r.Method == http.MethodPost && endpoint == "transacoes":
		defer r.Body.Close()
		err := json.NewDecoder(r.Body).Decode(&transacao)

		if err != nil {
			w.WriteHeader(400)
			return
		}

		err = validate.Struct(transacao)

		if err != nil {
			log.Println(err)
			w.WriteHeader(400)
			return
		}

		if _, err := db.Exec(
			"INSERT INTO TRANSACAO (VALOR, TIPO, DESCRICAO, CLIENTE, REALIZADA_EM) VALUES ($1, $2, $3, $4, $5)",
			transacao.Valor,
			transacao.Tipo,
			transacao.Descricao,
			clienteId,
			time.Now().Format("2006-01-02T03:04:05.000000Z")); err != nil {
			log.Println(err)
			return
		}

		if transacao.Tipo == "d" {
			transacao.Valor = -transacao.Valor
		}

		if _, err := db.Exec(
			"UPDATE CLIENTE SET SALDO = SALDO + $1 WHERE ID = $2",
			transacao.Valor,
			clienteId); err != nil {
			if err.Error() == "CHECK constraint failed: SALDO >= -LIMITE" {
				w.WriteHeader(422)
				return
			}

			log.Println(err)
			w.WriteHeader(500)
			return
		}

		row := db.QueryRow("SELECT LIMITE, SALDO FROM CLIENTE WHERE ID = $1", clienteId)

		cliente := Cliente{}

		err = row.Scan(&cliente.Limite, &cliente.Saldo)

		if err != nil {
			if err == sql.ErrNoRows {
				w.WriteHeader(404)
				return
			}

			log.Println(err)
			w.WriteHeader(500)
			return
		}

		if err := row.Err(); err != nil {
			log.Println(err)
			w.WriteHeader(500)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		err = json.NewEncoder(w).Encode(cliente)

		if err != nil {
			log.Println(err)
			w.WriteHeader(500)
			return
		}
	case r.Method == http.MethodGet && endpoint == "extrato":
		extrato := Extrato{}
		extrato.Saldo.DataExtrato = time.Now().Format("2006-01-02T03:04:05.000000Z")
		extrato.UltimasTransacoes = []Transacao{}

		row := db.QueryRow("SELECT LIMITE, SALDO FROM CLIENTE WHERE ID = $1", clienteId)

		err := row.Scan(&extrato.Saldo.Limite, &extrato.Saldo.Total)

		if err != nil {
			log.Println(err)
			w.WriteHeader(500)
			return
		}

		if err := row.Err(); err != nil {
			log.Println(err)
			w.WriteHeader(500)
			return
		}

		rows, err := db.Query("SELECT VALOR, TIPO, DESCRICAO, REALIZADA_EM FROM TRANSACAO WHERE CLIENTE = $1 ORDER BY REALIZADA_EM DESC LIMIT 10", clienteId)

		if err != nil {
			log.Println(err)
			w.WriteHeader(500)
			return
		}

		defer rows.Close()

		for rows.Next() {
			var transacao Transacao

			if err := rows.Scan(
				&transacao.Valor,
				&transacao.Tipo,
				&transacao.Descricao,
				&transacao.RealizadaEm); err != nil {
				if err == sql.ErrNoRows {
					break
				} else {
					log.Println(err)
					w.WriteHeader(500)
					return
				}
			}

			extrato.UltimasTransacoes = append(extrato.UltimasTransacoes, transacao)
		}

		if err := row.Err(); err != nil {
			log.Println(err)
			w.WriteHeader(500)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		err = json.NewEncoder(w).Encode(extrato)

		if err != nil {
			log.Println(err)
			w.WriteHeader(500)
			return
		}
	default:
		return
	}
}

func main() {
	var err error
	db, err = sql.Open("sqlite3", dburl)

	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()

	rotaClientes := "/clientes/"
	mux.Handle(rotaClientes, http.StripPrefix(rotaClientes, http.HandlerFunc(Handler)))

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      mux,
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	err = server.ListenAndServe()

	if err != nil {
		panic(fmt.Sprintf("cannot start server: %s", err))
	}
}
