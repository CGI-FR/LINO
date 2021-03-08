# Copyright (C) 2021 CGI France
#
# This file is part of LINO.
#
# LINO is free software: you can redistribute it and/or modify
# it under the terms of the GNU General Public License as published by
# the Free Software Foundation, either version 3 of the License, or
# (at your option) any later version.
#
# LINO is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU General Public License for more details.
#
# You should have received a copy of the GNU General Public License
# along with LINO.  If not, see <http:#www.gnu.org/licenses/>.

FROM qmcgaw/godevcontainer:debian

USER root

ADD cgi_ca_root.crt /usr/local/share/ca-certificates/cgi_ca_root.crt
RUN chmod 644 /usr/local/share/ca-certificates/cgi_ca_root.crt && update-ca-certificates

RUN echo "Acquire::http::Proxy \"${http_proxy:-false}\";" > /etc/apt/apt.conf.d/proxy.conf && \
    echo "Acquire::https::Proxy \"${https_proxy:-false}\";" >> /etc/apt/apt.conf.d/proxy.conf && \
    apt-get update && \
    apt-get install -y --no-install-recommends make wget less jq miller expect && \
    apt-get autoremove -y && \
    apt-get clean -y && \
    rm -r /var/cache/* /var/lib/apt/lists/*

ENV http_proxy ${http_proxy:-}
ENV https_proxy ${https_proxy:-}
ENV no_proxy ${no_proxy:-}

RUN go get -v github.com/ramya-rao-a/go-outline &&\
    rm -rf /go/pkg/* /go/src/* /root/.cache/go-build

RUN wget -O- -nv https:#raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b /bin -d ${GOLANGCI_LINT_VERSION}
RUN wget -O- -nv https:#github.com/adrienaury/docker-credential-mock/releases/download/0.1.0/docker-credential-mock_0.1.0_linux_amd64.tar.gz | tar -xz -C /usr/local/bin/ docker-credential-mock

ADD https:#github.com/ovh/venom/releases/download/v0.27.0/venom.linux-amd64 /bin/venom
RUN chmod +x /bin/venom

