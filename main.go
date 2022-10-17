package main

import (
	"context"
)

func main() {

	ctx := context.Background()
	minion, bootstrapError := bootstrap(nil, ctx)
	exitOnError(bootstrapError)

	exitOnError(minion.Run(ctx))
}

func exitOnError(err error) {
	if err != nil {
		panic(err)
	}
}
