
FROM golang:1.12.5 AS compiler

RUN apt-get update && apt-get install -y \
    xz-utils \
&& rm -rf /var/lib/apt/lists/*

ADD https://github.com/upx/upx/releases/download/v3.94/upx-3.94-amd64_linux.tar.xz /usr/local
RUN xz -d -c /usr/local/upx-3.94-amd64_linux.tar.xz | \
    tar -xOf - upx-3.94-amd64_linux/upx > /bin/upx && \
    chmod a+x /bin/upx

WORKDIR /src

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN GO111MODULE=on CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o build/sidevault .

RUN strip --strip-unneeded build/sidevault
RUN upx build/sidevault


FROM alpine:3.9.4

RUN apk --no-cache add ca-certificates && \
  update-ca-certificates

RUN mkdir -p /var/run/secrets/vaultproject.io/

COPY --from=compiler /src/build/sidevault /bin/sidevault

ENTRYPOINT [ "/bin/sidevault" ]
