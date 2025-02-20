package dns

import (
	"errors"
	"fmt"
)

var (
	ErrNoAvailableIPs = errors.New("no available IPs")
	ErrSubdomainInUse = errors.New("subdomain already in-use")
)

const (
	resolverFileInstructionFmt = `A custom DNS resolver is required to forward DNS requests to this server.
Create a file at %s with this content:

%s
`

	ipAliasInstruction = `One or more IP aliases are required for custom DNS routing.
Use the following commands to add an IPs:

  sudo ifconfig lo0 alias 10.0.0.1
  sudo ifconfig lo0 alias 10.0.0.2
  sudo ifconfig lo0 alias 10.0.0.3
  ...
`
)

func resolverFileInstructions(fname, expected string) string {
	return fmt.Sprintf(resolverFileInstructionFmt, fname, expected)
}

type UserFixableError struct {
	Err          error
	Instructions string
}

func NewUserFixableError(err error, instructions string) UserFixableError {
	return UserFixableError{
		Err:          err,
		Instructions: instructions,
	}
}

func (e UserFixableError) Error() string {
	return e.Err.Error()
}
