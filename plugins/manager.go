package plugins

import (
	"context"
	"fmt"
	"os"
	"plugin"
	"runtime"

	"github.com/calvinmclean/goblin/errors"
)

const (
	lookupErrorInstruction = `
The following function must be implemented in the main package:
    func Run(ctx context.Context, ipAddress string) error
`

	pluginErrorInstructionFmt = `
Go plugins are sensitive about versions.

Goblin was built with %s. Make sure to build your plugin with the same version.
If you recently updated Go, you should reinstall Goblin.

Goblin strives to have minimal dependencies, but it's possible your plugin has a different version of:
    - github.com/urfave/cli/v3
`
)

type RunFunc func(context.Context, string) error

func Load(fname string) (RunFunc, error) {
	_, err := os.Stat(fname)
	if err != nil {
		return nil, errors.NewUserFixableError(err, "\nDoes the file exist?\n")
	}

	p, err := plugin.Open(fname)
	if err != nil {
		return nil, errors.NewUserFixableError(err, fmt.Sprintf(pluginErrorInstructionFmt, runtime.Version()))
	}

	runSymb, err := p.Lookup("Run")
	if err != nil {
		return nil, errors.NewUserFixableError(err, lookupErrorInstruction)
	}

	runFunc, ok := runSymb.(func(context.Context, string) error)
	if !ok {
		return nil, errors.NewUserFixableError(fmt.Errorf("incorrect type: %T", runSymb), lookupErrorInstruction)
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
