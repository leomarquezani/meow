package schema

import (
	"time"
)

type Meow struct {
	Id        string    `json:"id"`
	Body      string    `json:"body"`
	CreatedAt time.Time `json:"createdAt"`
}
