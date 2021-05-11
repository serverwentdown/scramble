FROM golang:1.16-alpine3.13 AS build

RUN apk add \
	make

WORKDIR /go/src/app
COPY . .

RUN make TAGS=production


FROM alpine:3.13

RUN apk add --no-cache \
	tzdata

WORKDIR /app
COPY --from=build /go/src/app/scramble .

CMD ["./scramble"]
