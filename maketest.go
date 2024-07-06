//usage: go generate maketest.go

//go:build nope

package main

import "os"

//go:generate gofmt -e -d -s -w .
//go:generate go mod tidy
//go:generate go run $GOFILE
//go:generate go test -vet=all -coverprofile=bin/coverage/coverage_$GOOS.data -trimpath  ./...
//go:generate go tool cover -html=bin/coverage/coverage_$GOOS.data -o bin/coverage/coverage_$GOOS.html

func main() {
	// It's impossible (?) to prepare directory for coverage data with `go test` alone,
	// as `-coverprofile` attempts to write to disk before `-o` creates directory

	os.MkdirAll("bin/coverage", 0755)
}
