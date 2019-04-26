FROM golang:latest as builder
WORKDIR /go/src/github.com/ancientlore/hurl
ADD . .
RUN CGO_ENABLED=0 GOOS=linux GO111MODULE=on go get .
RUN CGO_ENABLED=0 GOOS=linux GO111MODULE=on go install

FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata
RUN addgroup -g 1000 hurl && adduser -G hurl -D -u 1000 hurl 
WORKDIR /home/hurl
COPY --from=builder /go/bin/hurl /usr/bin/hurl
EXPOSE 8080
USER hurl:hurl
ENTRYPOINT ["tail", "-f", "/dev/null"]
