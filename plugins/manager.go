package plugins

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"plugin"
	"runtime"

	"github.com/calvinmclean/goblin/errors"
)

const (
	lookupRunErrorInstruction = `
One of the following functions must be implemented in the main package:
    func Run(ctx context.Context, ipAddress string) error
    func Run(ctx context.Context) error // requires --env flag
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
		return nil, errors.NewUserFixableError(err, lookupRunErrorInstruction)
	}

	runFunc, ok := runSymb.(func(context.Context, string) error)
	if !ok {
		return nil, errors.NewUserFixableError(fmt.Errorf("incorrect type: %T", runSymb), lookupRunErrorInstruction)
	}

	return RunFunc(runFunc), nil
}

func LoadMainWithIPEnvVar(fname, ipEnvVar string) (RunFunc, error) {
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
		return nil, errors.NewUserFixableError(err, lookupRunErrorInstruction)
	}

	runFunc, ok := runSymb.(func(context.Context) error)
	if !ok {
		return nil, errors.NewUserFixableError(fmt.Errorf("incorrect type: %T", runSymb), lookupRunErrorInstruction)
	}

	return func(ctx context.Context, ipAddr string) error {
		err := os.Setenv(ipEnvVar, ipAddr)
		if err != nil {
			return fmt.Errorf("error setting IP env var: %w", err)
		}

		runFunc(ctx)
		return nil
	}, nil
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

// Build will use `go build -buildmode=plugin` to build a Plugin and return the path to the .so file
func Build(path string) (string, error) {
	originalDir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("failed to get current directory: %w", err)
	}

	defer func() {
		_ = os.Chdir(originalDir)
	}()

	err = os.Chdir(path)
	if err != nil {
		return "", fmt.Errorf("failed to chdir to plugin source: %w", err)
	}

	cmd := exec.Command("go", "build", "-buildmode=plugin")

	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", errors.NewUserFixableError(err, string(output))
	}

	pluginName := filepath.Base(path)

	return filepath.Join(path, pluginName) + ".so", nil
}
