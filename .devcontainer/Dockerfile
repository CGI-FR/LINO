FROM qmcgaw/godevcontainer:debian

USER root

RUN chown -R vscode:vscode /go

RUN apt-get update && \
    apt-get install -y --no-install-recommends make wget less && \
    apt-get autoremove -y && \
    apt-get clean -y && \
    rm -r /var/cache/* /var/lib/apt/lists/*

RUN go get -v github.com/ramya-rao-a/go-outline &&\
    rm -rf /go/pkg/* /go/src/* /root/.cache/go-build

RUN wget -O- -nv https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b /bin -d ${GOLANGCI_LINT_VERSION}

ADD https://github.com/ovh/venom/releases/download/v0.27.0/venom.linux-amd64 /bin/venom
RUN chmod +x /bin/venom

USER vscode