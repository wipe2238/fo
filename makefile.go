//usage: go generate makefile.go

//go:build nope

package main

//go:generate gofmt -e -d -s -w .
//go:generate go mod tidy -v
//go:generate go build -v -trimpath -o bin/ ./...
//go:generate go test -vet=all -trimpath ./...
