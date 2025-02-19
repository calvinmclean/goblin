package cmd

import (
	"context"
	"dns-plugin-thing/dns"
	"dns-plugin-thing/plugins"
	"fmt"
	"log"
	"time"
)

func runPlugin(ctx context.Context, dnsMgr dns.Manager, fname, hostname string, timeout time.Duration) error {
	if timeout != 0 {
		timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		ctx = timeoutCtx
	}

	log.Printf("starting plugin: %q", hostname)
	run, err := plugins.Load(fname)
	if err != nil {
		return fmt.Errorf("error loading plugin: %w", err)
	}

	ip, err := dnsMgr.GetIP(ctx, hostname)
	if err != nil {
		return fmt.Errorf("error getting IP: %w", err)
	}

	err = run(ctx, ip)
	log.Printf("stopped plugin: %q", hostname)
	return err
}
