FROM golang:1.8.3 as builder
WORKDIR /go/src/github.com/knarfeh/ruanyifeng2ebook/
COPY . /go/src/github.com/knarfeh/ruanyifeng2ebook/
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ruanyifeng2ebook .

FROM alpine:latest
WORKDIR /root/
COPY --from=builder /go/src/github.com/knarfeh/ruanyifeng2ebook/ruanyifeng2ebook /bin/
CMD ["ruanyifeng2ebook", "fetch"]