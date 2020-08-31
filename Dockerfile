FROM golang:alpine as builder 
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 go build

FROM scratch
WORKDIR /
COPY --from=builder ["/app/e-nr", "/app/dns.csv", "/app/index.html", "/"]
ENTRYPOINT [ "/e-nr" ]