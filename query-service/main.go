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
	"github.com/leomarquezani/meow/search"
)

type Config struct {
	PostgresDB           string `envconfig:"POSTGRES_DB"`
	PostgresUser         string `envconfig:"POSTGRES_USER"`
	PostgresPassword     string `envconfig:"POSTGRES_PASSWORD"`
	NatsAddress          string `envconfig:"NATS_ADDRESS"`
	ElasticsearchAddress string `envconfig:"ELASTICSEARCH_ADDRESS"`
}

func main() {

	var cfg Config
	err := envconfig.Process("", cfg)

	if err != nil {
		log.Fatal(err)
	}

	//connect to Postgres
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

	//connect to ElasticSearch
	retry.ForeverSleep(2*time.Second, func(_ int) error {
		addr := fmt.Sprintf("http://%s", cfg.ElasticsearchAddress)
		es, err := search.NewElastic(addr)

		if err != nil {
			log.Println(err)
			return err
		}
		search.SetRepository(es)
		return nil
	})
	defer search.Close()

	//connect to Nats
	retry.ForeverSleep(2*time.Second, func(_ int) error {
		addr := fmt.Sprintf("nats://%s", cfg.NatsAddress)
		nt, err := event.NewNats(addr)

		if err != nil {
			log.Println(err)
			return err
		}
		err = nt.OnMeowCreated(onMeowCreated)

		if err != nil {
			log.Println(err)
			return err
		}
		event.SetEventStore(nt)
		return nil
	})
	defer event.Close()

	router := NewRouter()
	if err := http.ListenAndServe(":6767", router); err != nil {
		log.Fatal(err)
	}
}

func NewRouter() (router *mux.Router) {
	router = mux.NewRouter()
	router.HandleFunc("/meows", listMeowsHandler).
		Methods("GET")
	router.HandleFunc("/search", searchMeowsHandler).
		Methods("GET")
	return
}
