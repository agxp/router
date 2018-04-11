FROM alpine:latest

RUN apk --no-cache add ca-certificates

RUN mkdir /app
WORKDIR /app
COPY ./bin/router .
COPY ./static/ ./static
CMD ["./router"]