package main

import (
	"context"

	"github.com/solumD/ozon-grapql-service/internal/app"
)

func main() {
	ctx := context.Background()
	app.InitAndRun(ctx)
}
