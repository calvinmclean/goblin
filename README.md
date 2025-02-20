# Title?

This program is used to run Go applications locally with DNS-resolved addresses. It works by running applications on private IPs and using a custom DNS resolver and server to access them.

## How does it work?

This consists of two main parts:

1. The server running in the background
    - This provides a DNS server for handling DNS resolution for your applications
    - It also runs a GRPC service that allows an application to request an IP and register a subdomain
2. Application wrapper
    - This component wraps a compiled Go plugin (`*.so` file). It handles the GRPC request to get an allocated IP and register the subdomain
    - This part is not strictly necessary since an application can be implemented to request an IP on its own. This method allows user applications to exist without any imports or specific handling related to domains and IPs


## Getting started

This requires a few system-level changes before it can be used. Eventually the server will handle these steps on its own if it's run with `sudo`, but for now it is manual setup since things generally should not run with `sudo`.

1. Create a custom top-level domain resolver setting at `/etc/resolver/{domain}` (The application's default is `gotest`)
    ```
    nameserver 127.0.0.1
    port 5053
    ```
    - If you create `/etc/resolver/gotest`, all DNS requests for `*.gotest` will use the DNS server at `127.0.0.1:5053`

1. Create IP aliases so your applications can run on private local IPs
    ```shell
    sudo ifconfig lo0 alias 10.0.0.1
    sudo ifconfig lo0 alias 10.0.0.2
    sudo ifconfig lo0 alias 10.0.0.3
    ...
    # create as many as you need
    sudo ifconfig lo0 alias 10.0.0.N
    ```
    - By default, the server expects to be able to use the `10.0.0.0/8` address block
    - These can be removed with:
        ```shell
        sudo ifconfig lo0 -alias 10.0.0.1
        ```

1. Run the server
    ```shell
    go run cmd/dns-plugin-thing/main.go server
    ```

1. Implement `Run(ctx context.Context, ipAddress string) error` in your application's `main` package and compile with `go build -buildmode=plugin` (or build the existing examples in this repository)
    ```shell
    task build-plugins
    # OR
    cd ./example-plugins/helloworld/cmd/hello && go build -buildmode=plugin
    cd ./example-plugins/helloworld/cmd/howdy && go build -buildmode=plugin
    ```

1. Run the plugin wrapper:
    ```shell
    go run cmd/dns-plugin-thing/main.go plugin -f ./example-plugins/helloworld/cmd/hello/hello.so -d hello
    ```

1. Use `curl` to make a request to the application using the registered domain name
    ```shell
    curl hello.gotest:8080
    ```

1. Repeat the last 2 steps with different subdomains and/or modules!


## About plugins

A [Go plugin](https://pkg.go.dev/plugin) is Go code compiled into a shared object (`.so) file that can be loaded and executed by another Go program at runtime. After loading a plugin, the program can look up a symbol by name and use type-assertion to use it like any other type. This means that the shared object file needs to provide the type that is expected.

This program could work without plugins by providing a wrapper library or `init` function that is imported by another application to handle the IP address allocation. Why didn't I do that instead?
- If this is a library, it would have to be imported into production code as well and controlled with other configurations. Since it exposes control to the program's operation externally, that's not ideal
- A library would likely have stricter implementation requirements and could interfere with normal development
- If the application is a plugin, this program has full control of its runtime and doesn't rely on user implementation
- As I add more implementation options that allow running different plugin symbols, this will be a really flexible and non-invasive way to execute a variety of applications
- This is an interesting plugins project and fun to learn about

It's definitely possible that this feature would be added optionally in the future, but for now it's an interesting plugins project.


## Implementing your own application

This program will look for and execute one of these symbols (in this order) from the plugin shared-object:

| Type                        | Symbol | Description          |
|-----------------------------|--------|----------------------|
| `func(ctx context.Context, ipAddress string) error` | `Run`   | The simple `Run` function can easily be used by a `main` function in a program's regular operation and also loaded by this wrapper |


### Example

This simple Hello World example shows how `Run(ctx context.Context, ipAddress string) error` can easily be implemented into a normal program.

```go
package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
)

func Run(ctx context.Context, ip string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "Hello, World!\n")
		log.Println("Hello, World!")
	})

	addr := fmt.Sprintf("%s:8080", ip)
	log.Printf("starting server on http://%s", addr)

	server := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	return server.ListenAndServe()
}

// allow this program to easily be run on its own
func main() {
	err := Run(context.Background(), "127.0.0.1")
	if err != nil {
		log.Fatal(err)
	}
}
```
