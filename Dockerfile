FROM golang:alpine as builder

RUN apk add --no-cache git

WORKDIR /build

COPY . .

RUN go get -d -v ./...

RUN CGO_ENABLED=0 go build -v -o go-monitoring main.go


FROM scratch

COPY --from=builder /build/go-monitoring /

USER 51862

EXPOSE 8080

ENTRYPOINT [ "/go-monitoring" ]
