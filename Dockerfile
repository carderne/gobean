FROM golang:1.21.5 AS builder
WORKDIR /build
COPY go.mod go.sum ./
RUN go mod download

COPY api api
COPY bean bean
COPY cmd cmd
COPY main.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -o ./gobean

FROM alpine:3.18.0 AS runner
WORKDIR /app
RUN apk add --no-cache ca-certificates
COPY --from=builder /build/gobean ./
ENV PORT=6767
EXPOSE $PORT
RUN adduser -D nonroot
USER nonroot
ENTRYPOINT [ "/app/gobean" ]
CMD ["api", "/file.bean"]
