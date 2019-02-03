FROM golang:1.11
WORKDIR /go/src/github.com/cnuber/dd-downtime

RUN go get github.com/golang/dep/cmd/dep
COPY Gopkg.toml Gopkg.lock ./
RUN dep ensure -v -vendor-only

COPY . ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /go/bin/dd-downtime ./src

FROM alpine:latest
RUN apk --no-cache add ca-certificates
COPY --from=0 /go/bin/dd-downtime /bin/dd-downtime
