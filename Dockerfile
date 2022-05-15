FROM golang:1.18 as builder

WORKDIR /app/source

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY ./ /app/source

ARG CGO_ENABLED=0
ARG GOOS=linux
ARG GOARCH=amd64

RUN mkdir /app/output
RUN go build -o /app/output ./cmd/...



FROM alpine

RUN apk add --no-cache ca-certificates tzdata && \
  cp /usr/share/zoneinfo/Asia/Tokyo /etc/localtime && \
  apk del tzdata
COPY --from=builder /app/output /apps
COPY cron_settings /root/

RUN crontab /root/cron_settings
CMD ["crond","-f","-d","8"]
