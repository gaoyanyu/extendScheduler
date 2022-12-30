FROM golang:1.18.5-alpine AS builder
MAINTAINER yanyu
WORKDIR /scheduler
COPY ./ /scheduler
RUN \
    GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -ldflags "-s -w" -o scheduler main.go && \
    chmod +x scheduler

FROM golang:1.18.5-alpine AS runner
WORKDIR /go/scheduler
COPY --from=builder /scheduler/scheduler .
COPY --from=builder /scheduler/secret/scheduler.conf /etc/scheduler/
CMD ["./scheduler"]