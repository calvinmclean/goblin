# TODO

This runs a simple TODO list API build with [babyapi](https://github.com/calvinmclean/babyapi).

This example implements `func Run(context.Context) error` and reads the IP address from IP_ADDR env var.

```shell
goblin run -p ./example-plugins/todo-api -d todo --env IP_ADDR
```

Then access with API with `curl` or `go run`:

```shell
cd example-plugins/todo-api/
go run main.go client --address "http://todo.goblin:8080" TODOs post --data '{"title": "use babyapi!"}'
go run main.go client --address "http://todo.goblin:8080" TODOs list

curl todo.goblin:8080/todos
```
