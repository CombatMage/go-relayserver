# STEP 1 build executable binary
FROM golang:1.13.0-alpine3.10 as builder

# Install git, required by go get
RUN apk update && apk add git

# Create appuser
RUN adduser -D -g '' appuser

# Copy app
COPY *.go /app/
COPY go.mod /app
COPY go.sum /app
WORKDIR /app

# build the binary, make sure that it is linked static
RUN CGO_ENABLED=0 GOOS=linux go build -o /go/bin/server


# STEP 2 build a small image
FROM scratch

# Copy appuser
COPY --from=builder /etc/passwd /etc/passwd

# Copy static executable from build
COPY --from=builder /go/bin/server /server

USER appuser
CMD  ["/server", \
    "-port-stream", "8000", \
    "-port-ws", "9000", \
    "-s", "secret1234"]
