FROM golang:1.15.3 as builder

WORKDIR /go/src

COPY go.mod go.sum ./
RUN go mod download

COPY . ./

ARG CGO_ENABLED=0
ARG GOOS=linux
RUN go build -o /go/bin/main -ldflags '-s -w'

FROM scratch as runner

COPY --from=builder /go/bin/main /app/main

ENTRYPOINT [ "/app/main" ]
