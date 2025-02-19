package main

import (
	"context"

	"dns-plugin-thing/plugins/examples/helloworld"
)

func Run(ctx context.Context, ip string) error {
	return helloworld.Run(ctx, "Hello", ip)
}
