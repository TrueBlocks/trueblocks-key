FROM golang:1.21-alpine

WORKDIR /app
COPY . /app
EXPOSE 5555

RUN cd queue/insert && go mod download  && cd -
ENTRYPOINT go run ./queue/insert/cmd
