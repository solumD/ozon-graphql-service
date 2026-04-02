package main

import (
	"context"

	"github.com/solumD/ozon-graphql-service/internal/app"
)

func main() {
	ctx := context.Background()
	app.InitAndRun(ctx)
}
