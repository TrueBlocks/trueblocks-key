# Move this Dockerfile to the root directory to use it

FROM golang:1-alpine AS builder

ADD . /src
WORKDIR /src
RUN mkdir bin
RUN go build -o bin/extract ./extract/

FROM alpine:3
COPY --from=builder /src/bin/extract /bin
ENTRYPOINT extract convert --database default /index
