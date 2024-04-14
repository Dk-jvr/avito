FROM golang:alpine

WORKDIR /avito

COPY . .

RUN go get -d -v ./...

RUN go install -v ./...

RUN go build -o /build

EXPOSE 8080

CMD ["/build"]
