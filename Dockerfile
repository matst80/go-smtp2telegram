# syntax=docker/dockerfile:1

FROM golang:alpine AS build-stage
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY *.go ./
#COPY views ./views
RUN CGO_ENABLED=0 GOOS=linux go build -o /go-crapmail

FROM gcr.io/distroless/base-debian11 
WORKDIR /

EXPOSE 8080

COPY --from=build-stage /go-crapmail /go-crapmail
ENTRYPOINT ["/go-crapmail"]