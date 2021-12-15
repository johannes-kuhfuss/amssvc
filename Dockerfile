# syntax=docker/dockerfile:1

##
## Build
##
FROM golang:1.17-alpine AS build

# Setup ENV
WORKDIR /app

# Download prereqs
COPY go.mod .
COPY go.sum .
RUN go mod download

# Copy sources
COPY . .

# Build
RUN go build -o /amssvc

##
## Deploy
##
FROM alpine:latest

WORKDIR /

COPY --from=build /amssvc /amssvc

EXPOSE 8080

USER nobody:nogroup

ENTRYPOINT ["/amssvc"]
