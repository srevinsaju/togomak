FROM golang:alpine AS builder 

WORKDIR /app
COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-w -s" -o /go/bin/togomak ./cmd/togomak 


FROM docker:cli AS docker 
RUN apk add --no-cache git 
COPY --from=builder /go/bin/togomak /usr/bin/togomak
ENTRYPOINT ["/usr/bin/togomak"]


FROM scratch AS tiny
COPY --from=builder /go/bin/togomak /usr/bin/togomak 
ENTRYPOINT ["/usr/bin/togomak"]

FROM alpine as alpine 
RUN apk add --no-cache git 
COPY --from=builder /go/bin/togomak /usr/bin/togomak 
ENTRYPOINT ["/usr/bin/togomak"]

FROM gcr.io/cloud-builders/docker AS docker-buster 
COPY --from=builder /go/bin/togomak /usr/bin/togomak 
ENTRYPOINT ["/usr/bin/togomak"]

FROM alpine as default
