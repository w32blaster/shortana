#
# Step 1: compile the app
#
FROM golang as builder

WORKDIR /app
COPY . .

# Run tests
RUN make test

# compile app
RUN CGO_ENABLED=0 GOOS=linux go build \
    -a -installsuffix cgo \
    -ldflags "-s -w" \
    -o /app/bot \
    cmd/shortana/main.go

#
# Phase 2: prepare the runtime container, ready for production
#
FROM scratch

VOLUME "/storage"
EXPOSE 8444

# copy our bot executable
COPY --from=builder /app/bot /bot

# copy root CA certificate to set up HTTPS connection with Telegram
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

CMD ["/bot"]