package plugins

import (
	"context"
	"fmt"
	"plugin"
)

type RunFunc func(context.Context, string) error

func Load(fname string) (RunFunc, error) {
	p, err := plugin.Open(fname)
	if err != nil {
		return nil, err
	}

	runSymb, err := p.Lookup("Run")
	if err != nil {
		return nil, err
	}

	runFunc, ok := runSymb.(func(context.Context, string) error)
	if !ok {
		return nil, fmt.Errorf("incorrect type: %T", runSymb)
	}

	return RunFunc(runFunc), nil
}

type IPGetter interface {
	GetIP(ctx context.Context, subdomain string) (string, error)
}

func Run(ctx context.Context, run RunFunc, getter IPGetter, subdomain string) error {
	ip, err := getter.GetIP(ctx, subdomain)
	if err != nil {
		return fmt.Errorf("error getting IP: %w", err)
	}

	return run(ctx, ip)
}
