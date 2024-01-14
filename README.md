# gobean

Basic [beancount](https://github.com/beancount/beancount) clone (one day...) written in Go.

I'm deliberately writing this without looking at the beancount source code, or trying too hard to build a proper lexer or AST (because I'm trying to have fun).

Go is a terrible choice for this particular problem (limited switch/pattern matching, and no enum/union types) so I'm not sure how far I'll go with this...

Planned features:
- [x] Parse beancount files
- [x] Calculate account balances
- [x] Use Decimal
- [x] Propagate line numbers for debugging
- [x] Stricter transaction keywords
- [x] Price directives
- [x] Pad directives
- [x] Validate transactions against `open`/`close` directives
- [ ] Validate `balance` directives
- [ ] Open/close with multiple curencies

## Usage
### Install
```bash
go install github.com/carderne/gobean@latest
```

### Run
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
### Install dependencies
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
