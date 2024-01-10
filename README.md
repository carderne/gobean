# gobean

Basic [beancount](https://github.com/beancount/beancount) clone (one day...) written in Go.

Planned features:
- [x] Parse beancount files
- [x] Calculate account balances
- [x] Use Decimal
- [x] Propagate line numbers for debugging
- [ ] Validate against `open`/`close` directives
- [ ] Open/close with multiple curencies
- [ ] Validate against `balance` directives
- [ ] Price directives
- [ ] Pad directives

## Usage
Install:
```bash
go install github.com/carderne/gobean@latest
```

Run:
```
$ gobean

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

## Development
Install deps:
```bash
git clone git@github.com:carderne/gobean.git
cd gobean
go get .
```

### Build
Requires at least Go `v1.21`.

```bash
# automatically fmts and vets
make
```

### Lint
```bash
make lint
```

### Test
```bash
make test
```

## Docker

```bash
docker build --tag carderne/gobean .
docker run -p 6767:6767 --rm -v ./example.bean:/file.bean carderne/gobean
```
