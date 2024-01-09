# gobean

Basic [beancount](https://github.com/beancount/beancount) clone (one day...) written in Go.

Planned features:
- [x] Parse beancount files
- [x] Calculate account balances
- [ ] Use [shopspring/decimal](https://github.com/shopspring/decimal)
- [ ] Propagate line numbers for debugging
- [ ] Validate against `open`/`close` directives
- [ ] Open/close with multiple curencies
- [ ] Validate against `balance` directives
- [ ] Price directives
- [ ] Pad directives

## Usage
```
$ go run .

NAME:
   gobean - A new cli application

USAGE:
   gobean [global options] command [command options]

COMMANDS:
   api, a       Run the API for a beancount file
   balances, v  Print all account balances
   help, h      Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --help, -h  show help
```

## Linting etc
```bash
go vet ./...
go fmt ./...
golint ./...

## Test
```bash
go test
```

## Build
Requires at least Go `v1.21`.

```bash
# automatically fmts and vets
make
```


## Docker (why?)

```bash
docker build --tag carderne/gobean .
docker run -p 6767:6767 --rm -v ./example.bean:/file.bean carderne/gobean
```
