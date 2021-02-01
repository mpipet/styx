FROM golang

WORKDIR /go/src/gitlab.com/dataptive/styx

COPY . .

RUN go get -d -v ./cmd/...
RUN go install -v ./cmd/...

ENTRYPOINT styx-server

EXPOSE 8000
