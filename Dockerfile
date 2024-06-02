FROM golang:1.22-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
RUN go build -ldflags="-s -w" -o wsserver .

FROM scratch

COPY --from=builder ["/build/wsserver","/"]

EXPOSE 3000
EXPOSE 4040

ENTRYPOINT ["/wsserver"]