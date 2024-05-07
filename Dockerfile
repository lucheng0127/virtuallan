FROM golang:1.20 as builder

ARG TARGETOS
ARG TARGETARCH
ENV GOPROXY=https://goproxy.io,direct

WORKDIR /

COPY go.mod go.mod
COPY go.sum go.sum
COPY pkg pkg
COPY main.go main.go

RUN CGO_ENABLED=1 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} go build -a -o virtuallan main.go

FROM alpine:3.19

WORKDIR /

RUN sed -i 's/dl-cdn.alpinelinux.org/mirrors.ustc.edu.cn/g' /etc/apk/repositories
RUN apk add --no-cache gcompat

COPY config config
COPY static static
COPY --from=builder virtuallan virtuallan
CMD /virtuallan server -d /config