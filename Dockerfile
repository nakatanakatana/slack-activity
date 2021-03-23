FROM golang:1.16 as builder

WORKDIR /apps
COPY . /apps
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o joinAllChannels cmd/joinAllChannels/main.go
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o reportUnusedChannels cmd/reportUnusedChannels/main.go

FROM alpine

RUN apk add --no-cache ca-certificates tzdata && \
  cp /usr/share/zoneinfo/Asia/Tokyo /etc/localtime && \
  apk del tzdata
COPY --from=builder /apps/joinAllChannels /
COPY --from=builder /apps/reportUnusedChannels /
COPY cron_settings /root/

RUN crontab /root/cron_settings
CMD ["crond","-f","-d","8"]
