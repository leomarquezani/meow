package main

import (
	"html/template"
	"log"
	"net/http"
	"time"

	"github.com/leomarquezani/meow/event"

	"github.com/leomarquezani/meow/schema"

	"github.com/leomarquezani/meow/db"
	"github.com/leomarquezani/meow/util"
	"github.com/segmentio/ksuid"
)

func createMeowHandler(w http.ResponseWriter, r *http.Request) {
	type response struct {
		Id string `json:"id"`
	}

	ctx := r.Context()

	//Read parameters
	body := template.HTMLEscapeString(r.FormValue("body"))
	if len(body) < 1 || len(body) > 140 {
		util.ResponseError(w, http.StatusBadRequest, "Invalid Body")
		return
	}

	//Create Meow
	createdAt := time.Now().UTC()
	id, err := ksuid.NewRandomWithTime(createdAt)
	if err != nil {
		util.ResponseError(w, http.StatusInternalServerError, "Failed to create meow")
		return
	}

	meow := schema.Meow{
		Id:        id.String(),
		Body:      body,
		CreatedAt: createdAt,
	}

	if err := db.InsertMeow(ctx, meow); err != nil {
		log.Println(err)
		util.ResponseError(w, http.StatusInternalServerError, "Failed to create meow")
		return
	}

	//publish event
	if err := event.PublishMeowCreated(meow); err != nil {
		log.Println(err)
	}

	//return new Meow
	util.ResponseOk(w, response{Id: meow.Id})

}
