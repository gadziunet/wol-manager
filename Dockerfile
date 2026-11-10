FROM --platform=$BUILDPLATFORM golang:1.24-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY *.go ./
COPY web ./web
ARG VERSION=dev
ARG TARGETOS
ARG TARGETARCH
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH go build -trimpath -ldflags="-s -w -X main.version=${VERSION}" -o /out/wol . \
 && mkdir /out/data

FROM scratch
COPY --from=build --chown=65532:65532 /out/data /data
COPY --from=build /out/wol /wol
USER 65532:65532
ENV WOL_ADDR=:8080 WOL_DATA=/data
VOLUME ["/data"]
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 CMD ["/wol", "healthcheck"]
ENTRYPOINT ["/wol"]
