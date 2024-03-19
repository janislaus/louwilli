FROM golang:1.21.4  AS build
WORKDIR /app
COPY . .
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH  go build -o louie-admin

FROM gcr.io/distroless/base-debian11

ENV TZ="Europe/Berlin"
ENV GOPATH /usr/share

WORKDIR $GOPATH

COPY --from=build /app/louie-admin $GOPATH
COPY --from=build /app/admin/static $GOPATH/admin/static

USER nonroot:nonroot

ENTRYPOINT ["/usr/share/louie-admin"]