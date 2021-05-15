# dpkg --install

## Building

```sh
$ env GOOS=linux GOARCH=amd64 go build -o dpkg install.go
```

## Running

```sh
$ dpkg test.deb
```
