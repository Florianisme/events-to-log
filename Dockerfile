FROM golang:1.24-alpine as golang

RUN adduser \
  --disabled-password \
  --gecos "" \
  --home "/nonexistent" \
  --shell "/sbin/nologin" \
  --no-create-home \
  --uid 65532 \
  event-user

WORKDIR /app

COPY . .

RUN go mod download
RUN go mod verify

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -v -o /events-to-log .

FROM gcr.io/distroless/static-debian12

COPY --from=golang /events-to-log .

CMD ["./events-to-log"]