# dpkg --build

## Building

```sh
$ env GOOS=linux GOARCH=amd64 go build -o dpkg build.go
```

## Running

```sh
$ dpkg 1.1-1 test.deb
```
