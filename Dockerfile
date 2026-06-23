FROM golang:1.26-bookworm

RUN groupadd -g 1000 go && \
    useradd -u 1000 -g 1000 -d /home/go -m -s /bin/bash go

WORKDIR /home/go/app

RUN chown -R go:go /home/go/app

USER go
