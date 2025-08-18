FROM golang:1.24-alpine AS builder

RUN apk add --no-cache curl ca-certificates

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/bootstrap ./main.go \
 && curl -Lo /usr/local/bin/aws-lambda-rie https://github.com/aws/aws-lambda-runtime-interface-emulator/releases/latest/download/aws-lambda-rie \
 && chmod +x /usr/local/bin/aws-lambda-rie

FROM scratch

COPY --from=builder /app/bootstrap /var/task/bootstrap

COPY --from=builder /usr/local/bin/aws-lambda-rie /usr/local/bin/aws-lambda-rie

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

ENTRYPOINT ["/var/task/bootstrap"]
