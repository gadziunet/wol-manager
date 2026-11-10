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
# Starts as root only to hand /data to the unprivileged user, then switches to
# it (WOL_UID/WOL_GID, default 65532) before opening any network port.
ENV WOL_ADDR=:8080 WOL_DATA=/data
VOLUME ["/data"]
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 CMD ["/wol", "healthcheck"]
ENTRYPOINT ["/wol"]
