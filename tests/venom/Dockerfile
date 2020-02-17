# Build arguments
ARG IMAGE_NAME
ARG IMAGE_TAG
ARG IMAGE_REVISION
ARG IMAGE_DATE

FROM alpine AS builder

ADD https://github.com/ovh/venom/releases/download/v0.27.0/venom.linux-amd64 /venom
RUN chmod +x /venom

FROM gcr.io/distroless/base

COPY --from=builder /venom /

ENTRYPOINT [ "/venom" ]


# OCI labels (https://github.com/opencontainers/image-spec/blob/master/annotations.md)
LABEL org.opencontainers.image.created="${IMAGE_DATE}"
LABEL org.opencontainers.image.authors="Youen PERON <youen.peron@cgi.com>, Adrien AURY <adrien.aury@cgi.com>"
LABEL org.opencontainers.image.url="https://makeit.imfr.cgi.com/makeit2/scm/lino/lino-cli"
LABEL org.opencontainers.image.documentation="https://makeit.imfr.cgi.com/makeit2/scm/lino/lino-cli"
LABEL org.opencontainers.image.source="https://makeit.imfr.cgi.com/makeit2/scm/lino/lino-cli"
LABEL org.opencontainers.image.version="${IMAGE_TAG}"
LABEL org.opencontainers.image.revision="${IMAGE_REVISION}"
LABEL org.opencontainers.image.vendor="CGI"
LABEL org.opencontainers.image.licenses="UNLICENSED"
LABEL org.opencontainers.image.ref.name="${IMAGE_NAME}:${IMAGE_TAG}"
LABEL org.opencontainers.image.title="${IMAGE_NAME}"
LABEL org.opencontainers.image.description="Venom for Lino"
