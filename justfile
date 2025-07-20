[private]
default:
  @just --list

# build the application
build:
    @mkdir -p build/
    go build -o build/app

# run the application
run:
    go run .

# run both tests and linters
check: test lint

# run tests
test:
    go test -v ./...

# run linter
lint:
    golangci-lint run

# format code
fmt:
    golangci-lint fmt
