package db

import (
	"context"

	"github.com/leomarquezani/meow/schema"
)

type Repository interface {
	Close()
	InsertMeow(ctx context.Context, meow schema.Meow)
	ListMeows()
}
