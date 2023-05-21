# ENV is for zerolog
export ENV = dev

.DEFAULT_GOAL = build

.PHONY: build
build: fmt vet
	go build -v -o bin/gobean

.PHONY: fmt
fmt:
	go fmt ./...

.PHONY: lint
lint:
	golint ./...

.PHONY: vet
vet:
	go vet ./...

.PHONY: watch
watch:
	fd .go | entr -r go run . api example.bean

.PHONY: test
test:
	go test ./...
