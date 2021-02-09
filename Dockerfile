FROM golang AS builder

WORKDIR /go/src/gitlab.com/dataptive/styx

COPY . .

RUN go get -d -v ./cmd/...
RUN go install -v ./cmd/...
RUN CGO_ENABLED=0 go build -o $GOPATH/bin/styx ./cmd/styx 
RUN CGO_ENABLED=0 go build -o $GOPATH/bin/styx-server ./cmd/styx-server

FROM alpine:latest  
WORKDIR /etc/styx

COPY --from=builder /go/src/gitlab.com/dataptive/styx/config.toml config.toml
COPY --from=builder /go/bin /usr/bin

ENTRYPOINT ["styx-server", "--config", "/etc/styx/config.toml"]

EXPOSE 8000
