
build-hello:
	env GOOS=linux GOARCH=amd64 go build -o bin/hello hello.go

run-hello:
	go run hello.go