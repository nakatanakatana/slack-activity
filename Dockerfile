FROM alpine:latest

RUN apk --update upgrade && \
    apk add curl ca-certificates && \
    update-ca-certificates && \
    rm -rf /var/cache/apk/*

COPY slack-activity /
CMD ["/slack-activity" ]

