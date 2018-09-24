package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"github.com/kelseyhightower/envconfig"
	"github.com/leomarquezani/meow/db"
	"github.com/leomarquezani/meow/event"
	"github.com/leomarquezani/meow/retry"
)

func main() {

	var cfg Config

	err := envconfig.Process("", &cfg)

	if err != nil {
		log.Fatal(err)
	}

	retry.ForeverSleep(2*time.Second, func(attempt int) error {
		addr := fmt.Sprintf("postgres://%s:%s@postgres/%s?sslmode=disable", cfg.PostgresUser, cfg.PostgresPassword, cfg.PostgresDB)
		repo, err := db.NewPostgres(addr)

		if err != nil {
			log.Println(err)
			return err
		}
		db.SetRepository(repo)
		return nil
	})
	defer db.Close()

	retry.ForeverSleep(2*time.Second, func(_ int) error {
		es, err := event.NewNats(fmt.Sprintf("nats://%s", cfg.NatsAddress))
		if err != nil {
			log.Println(err)
			return nil
		}
		event.SetEventStore(es)
		return nil
	})
	defer event.Close()

	router := newRouter()

	if err := http.ListenAndServe(":6767", router); err != nil {
		log.Fatal(err)
	}
}

func newRouter() (router *mux.Router) {
	router = mux.NewRouter()
	router.HandleFunc("/meows", createMeowHandler).
		Methods("POST").
		Queries("body", "{body}")
	return
}

type Config struct {
	PostgresDB       string `envconfig:"POSTGRES_DB"`
	PostgresUser     string `envconfig:"POSTGRES_USER"`
	PostgresPassword string `envconfig:"POSTGRES_PASSWORD"`
	NatsAddress      string `envconfig:"NATS_ADDRESSxx"`
}
