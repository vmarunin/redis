FROM golang:1.19 AS build
WORKDIR /go/src
COPY . ./

ENV CGO_ENABLED=0
RUN go get -d -v ./...

RUN go build -a .
RUN go test

FROM scratch AS runtime
COPY --from=build /go/src/ ./
EXPOSE 8080/tcp
ENTRYPOINT ["./redis"]
