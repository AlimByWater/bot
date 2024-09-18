FROM golang:latest as builder
WORKDIR /app
COPY . .
RUN go get github.com/mailru/easyjson && go install github.com/mailru/easyjson/...@latest
RUN easyjson internal/entity/*.go

RUN apt-get update \
 && DEBIAN_FRONTEND=noninteractive \
    apt-get install --no-install-recommends --assume-yes \
      protobuf-compiler

RUN go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
RUN go install google.golang.org/protobuf/cmd/protoc-gen-go@latest

RUN go mod download && \
     protoc --go_out=. --go-grpc_out=. pkg/proto/downloader.proto && \
    CGO_ENABLED=0 GOOS=linux go build -o bot main.go


FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/bot /app/bot
COPY --from=builder /app/configs /app/configs
RUN mkdir /app/temp

EXPOSE 443
ENTRYPOINT ["./bot"]