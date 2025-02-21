# Fallback

## How To

1. Run the "remote" application (not managed by Goblin):
    ```shell
    cd example-plugins/fallback
    go run main.go
    ```

2. In another terminal, run Goblin server with the example fallback routes:
    ```shell
    goblin server -r example-fallback-routes.json
    ```

3. In another terminal, run the `fallback` application as a Goblin plugin:
    ```shell
    cd example-plugins/fallback
    go build -buildmode=plugin

    goblin plugin -f ./fallback.so -d fallback
    ```

4. In another terminal, use `curl` with the Goblin subdomain:
    ```shell
    > curl fallback.goblin:8081
    I am running at 10.0.0.1:8081
    ```

5. In the terminal from step 3, use Ctrl-C to stop the plugin

6. Use the same `curl` command from step 4 to see how Goblin routes to the service running locally from `main.go`:
    ```shell
    > curl fallback.goblin:8081
    I am running at 0.0.0.0:8081
    ```
