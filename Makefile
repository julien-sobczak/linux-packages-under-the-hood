
build-hello:
	env GOOS=linux GOARCH=amd64 go build -o bin/hello hello.go
run-hello:
	go run hello.go

test:
	go test ./...

build-cmd:
	env GOOS=linux GOARCH=amd64 go build -o bin/dpkg cmd/dpkg/*.go
	env GOOS=linux GOARCH=amd64 go build -o bin/apt cmd/apt/*.go
