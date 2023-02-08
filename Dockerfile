FROM golang:1.20-alpine as builder

RUN apk add -U --no-cache ca-certificates
RUN apk add -U git tzdata upx

ENV SRC_DIR=/build/

WORKDIR $SRC_DIR

RUN go env -w GOSUMDB=off
COPY go.* ./
RUN CGO_ENABLED=0 go mod download

COPY . $SRC_DIR

RUN CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go test ./...
RUN GIT_BRANCH=$(git branch | grep \* | cut -d ' ' -f2); \
    GIT_HASH=$(git rev-parse --short HEAD); \
    GIT_DIRTY=$(if git diff --quiet; then echo false; else echo true; fi;); \
    BUILD_DATE=$(date +%Y.%m.%d_%H:%M:%S); \
    CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -ldflags "-s -w" -o /bin/morphbits -a ./cmd
RUN upx --best --lzma /bin/morphbits

FROM scratch
COPY --from=builder /bin/morphbits /bin/morphbits
COPY --from=builder /usr/share/zoneinfo /usr/share/zoneinfo
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY ./data /etc/morphbits/data

EXPOSE 9090
ENTRYPOINT ["/bin/morphbits"]
