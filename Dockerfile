FROM golang:1.16-alpine3.13 AS build

WORKDIR /go/src/app
COPY . .

RUN go build


FROM alpine:3.13

WORKDIR /app
COPY --from=build /go/src/app/scramble .

CMD ["./scramble"]
