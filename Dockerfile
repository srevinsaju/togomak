FROM golang:1.19 AS builder
WORKDIR /app
COPY . .
RUN cd cmd/togomak && go build -o /togomak


FROM gcr.io/cloud-builders/docker 
COPY --from=builder /togomak /app/togomak
ENTRYPOINT ["/app/togomak"]

