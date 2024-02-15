FROM golang:1.22-alpine

RUN apk add lsof

WORKDIR /app
COPY . /app
EXPOSE 5555

RUN cd queue/insert && go mod download  && cd -
ENTRYPOINT go run ./queue/insert/cmd
