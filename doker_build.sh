#! /bin/bash

CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o slack-activity
docker build . -t nakatanakatana/slack-activity
