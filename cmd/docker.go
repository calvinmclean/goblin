package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"

	"github.com/calvinmclean/goblin/dns"
	"github.com/calvinmclean/goblin/errors"

	"github.com/urfave/cli/v3"
)

const defaultDockerSocket = "/var/run/docker.sock"

var (
	dockerSocketEnvVar = cli.EnvVar("DOCKER_SOCK")

	dockerContainer, dockerSocket string
	DockerCmd                     = &cli.Command{
		Name:        "docker",
		Description: "register a docker container with a subdomain",
		Action:      runRegisterDocker,
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "port",
				Value:       defaultServerPort,
				Usage:       "port to reach the API server running locally",
				Destination: &serverPort,
				Sources:     cli.ValueSourceChain{Chain: []cli.ValueSource{portEnvVar}},
			},
			&cli.StringFlag{
				Name:        "subdomain",
				Aliases:     []string{"d"},
				Usage:       "subdomain name",
				DefaultText: "docker container name (if using --container/-c)",
				Destination: &subdomain,
			},
			&cli.StringFlag{
				Name:        "container",
				Aliases:     []string{"c"},
				Usage:       "name of a docker container running locally",
				Destination: &dockerContainer,
				Required:    true,
			},
			&cli.StringFlag{
				Name:        "socket",
				Usage:       "path to the docker socket",
				Destination: &dockerSocket,
				Value:       defaultDockerSocket,
				Sources:     cli.ValueSourceChain{Chain: []cli.ValueSource{dockerSocketEnvVar}},
			},
		},
	}
)

func runRegisterDocker(ctx context.Context, c *cli.Command) error {
	if subdomain == "" {
		subdomain = dockerContainer
	}

	containerIP, err := getContainerIP(dockerContainer)
	if err != nil {
		return fmt.Errorf("error getting IP for container: %w", err)
	}

	client, err := dns.NewHTTPClient(net.JoinHostPort(defaultAddr, serverPort))
	if err != nil {
		return fmt.Errorf("error creating client: %w", err)
	}

	err = client.RegisterFallback(subdomain, containerIP)
	if err != nil {
		return fmt.Errorf("error registering docker container: %w", err)
	}
	log.Print("registered docker container")

	return err
}

func getContainerIP(containerName string) (string, error) {
	url := fmt.Sprintf("http://localhost/containers/%s/json", containerName)
	req, err := http.NewRequest(http.MethodGet, url, http.NoBody)
	if err != nil {
		return "", fmt.Errorf("error creating request: %w", err)
	}

	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
				return net.Dial("unix", dockerSocket)
			},
		},
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %w", err)
	}

	var containerData struct {
		NetworkSettings struct {
			Networks map[string]struct {
				IPAddress string `json:"IPAddress"`
			} `json:"Networks"`
		} `json:"NetworkSettings"`
	}
	if err := json.Unmarshal(body, &containerData); err != nil {
		return "", fmt.Errorf("error parsing response: %w", err)
	}

	for _, network := range containerData.NetworkSettings.Networks {
		return network.IPAddress, nil
	}

	return "", errors.New("no IP address found for container")
}
