FROM golang:1.13 AS builder

ADD .devcontainer/cgi_ca_root.crt /usr/local/share/ca-certificates/cgi_ca_root.crt
RUN chmod 644 /usr/local/share/ca-certificates/cgi_ca_root.crt && update-ca-certificates

ENV GOFLAGS="-mod=readonly"

RUN mkdir /home/lino

RUN mkdir -p /workspace
WORKDIR /workspace

ARG GOPROXY

COPY go.* /workspace/
RUN go mod download

COPY . /workspace

ARG VERSION
ARG BUILD_BY

RUN make release

FROM gcr.io/distroless/base

# Build arguments
ARG IMAGE_NAME
ARG IMAGE_TAG
ARG IMAGE_REVISION
ARG IMAGE_DATE

# OCI labels (https://github.com/opencontainers/image-spec/blob/master/annotations.md)
LABEL org.opencontainers.image.created="${IMAGE_DATE}"
LABEL org.opencontainers.image.authors="Youen PERON <youen.peron@cgi.com>, Adrien AURY <adrien.aury@cgi.com>"
LABEL org.opencontainers.image.url="https://makeit.imfr.cgi.com/makeit2/scm/perony/lino"
LABEL org.opencontainers.image.documentation="https://makeit.imfr.cgi.com/makeit2/scm/perony/lino"
LABEL org.opencontainers.image.source="https://makeit.imfr.cgi.com/makeit2/scm/perony/lino"
LABEL org.opencontainers.image.version="${IMAGE_TAG}"
LABEL org.opencontainers.image.revision="${IMAGE_REVISION}"
LABEL org.opencontainers.image.vendor="CGI"
LABEL org.opencontainers.image.licenses="UNLICENSED"
LABEL org.opencontainers.image.ref.name="${IMAGE_NAME}:${IMAGE_TAG}"
LABEL org.opencontainers.image.title="${IMAGE_NAME}"
LABEL org.opencontainers.image.description="Lino is a simple ETL (Extract Transform Load) tools to manage tests datas."

COPY --from=builder /home/lino /home/lino
COPY --from=builder /workspace/bin/* /

WORKDIR /home/lino

ENTRYPOINT [ "/lino" ]

