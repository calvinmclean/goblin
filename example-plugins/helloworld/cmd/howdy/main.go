package main

import (
	"context"

	"helloworld"
)

func Run(ctx context.Context, ip string) error {
	return helloworld.Run(ctx, "Howdy", ip)
}
