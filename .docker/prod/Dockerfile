FROM golang:1.8.3 as builder
WORKDIR /go/src/github.com/ee-book/ruanyifengeebook/
COPY . /go/src/github.com/ee-book/ruanyifengeebook/
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ruanyifengeebook.

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /go/src/github.com/ee-book/ruanyifengeebook/ruanyifengeebook /bin/
CMD ["ruanyifengeebook", "fetch"]