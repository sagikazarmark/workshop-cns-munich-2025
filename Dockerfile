FROM --platform=$BUILDPLATFORM golang:1.24-alpine3.21 AS builder

RUN apk add --update --no-cache ca-certificates git

ARG TARGETOS
ARG TARGETARCH
ARG TARGETPLATFORM

WORKDIR /usr/local/src/app

ARG GOPROXY

ENV CGO_ENABLED=0
ENV GOOS=$TARGETOS GOARCH=$TARGETARCH

COPY go.* ./
RUN go mod download

COPY . .

ARG VERSION

RUN go build -trimpath -ldflags "-X main.version=${VERSION}" -o /usr/local/bin/app .

FROM alpine:3.21

RUN apk add --update --no-cache ca-certificates tzdata

COPY --from=builder /usr/local/bin/app /usr/local/bin/

EXPOSE 8080

CMD ["app"]
