# TODO

This runs a simple TODO list API build with [babyapi](https://github.com/calvinmclean/babyapi).

```shell
cd ./example-plugins/todo-api
go build -buildmode=plugin

cd -
go run cmd/dns-plugin-thing/main.go plugin -f ./example-plugins/todo-api/todo-api.so -d todo
```

Then access with API with `curl` or `go run`:

```shell
cd example-plugins/todo-api/
go run main.go client --address "http://todo.gotest:8080" TODOs post --data '{"title": "use babyapi!"}'
go run main.go client --address "http://todo.gotest:8080" TODOs list

curl todo.gotest:8080/todos
```
