FROM golang:alpine AS build-service
COPY src /tmp/src
WORKDIR /tmp/src
RUN mkdir -p /tmp/build
RUN go mod download
RUN go build -o /tmp/build/app

FROM alpine:latest
COPY --from=build-service /tmp/build/app /service
COPY res /res
RUN apk --no-cache add curl
ENTRYPOINT ["/service"]
EXPOSE 8000
# HEALTHCHECK --interval=5s CMD curl -s -f http://localhost:8000/healthcheck
