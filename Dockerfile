FROM golang:latest as builder
WORKDIR /app
COPY . .
RUN go get github.com/mailru/easyjson && go install github.com/mailru/easyjson/...@latest
RUN easyjson internal/entity/layout.go
RUN go mod download && CGO_ENABLED=0 GOOS=linux go build -o bot main.go


FROM alpine:latest
WORKDIR /app
COPY --from=builder /app/bot /app/bot
COPY --from=builder /app/configs /app/configs
EXPOSE 443
ENTRYPOINT ["./bot"]