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
# along with LINO.  If not, see <http://www.gnu.org/licenses/>.

FROM adrienaury/go-devcontainer:v0.6-debian

USER root

ADD cgi_ca_root.crt /usr/local/share/ca-certificates/cgi_ca_root.crt
RUN chmod 644 /usr/local/share/ca-certificates/cgi_ca_root.crt && update-ca-certificates

RUN echo "Acquire::http::Proxy \"${http_proxy:-false}\";" > /etc/apt/apt.conf.d/proxy.conf && \
    echo "Acquire::https::Proxy \"${https_proxy:-false}\";" >> /etc/apt/apt.conf.d/proxy.conf && \
    apt-get update && \
    apt-get install -y --no-install-recommends make less jq expect libaio1 wget unzip gcc-mingw-w64 g++-mingw-w64 gcc-multilib gcc-mingw-w64 libxml2-dev && \
    apt-get autoremove -y && \
    apt-get clean -y && \
    rm -r /var/cache/* /var/lib/apt/lists/*

ENV http_proxy ${http_proxy:-}
ENV https_proxy ${https_proxy:-}
ENV no_proxy ${no_proxy:-}

ENV DEBIAN_FRONTEND noninteractive

# Oracle
RUN wget -O /tmp/instantclient-basic-linux-x64.zip https://download.oracle.com/otn_software/linux/instantclient/193000/instantclient-basic-linux.x64-19.3.0.0.0dbru.zip && \
    mkdir -p /usr/lib/oracle && \
    unzip /tmp/instantclient-basic-linux-x64.zip -d /usr/lib/oracle && \
    ldconfig -v /usr/lib/oracle/instantclient_19_3 && \
    ldd /usr/lib/oracle/instantclient_19_3/libclntsh.so

ARG VERSION_DCM=0.1.0
RUN wget -O- -nv https://github.com/adrienaury/docker-credential-mock/releases/download/${VERSION_DCM}/docker-credential-mock_${VERSION_DCM}_linux_amd64.tar.gz | tar -xz -C /usr/local/bin/ docker-credential-mock \
    && chmod +x /usr/local/bin/docker-credential-mock

ARG VERSION_YAML2JSON=1.3
ADD https://github.com/bronze1man/yaml2json/releases/download/v${VERSION_YAML2JSON}/yaml2json_linux_amd64 /usr/bin/yaml2json
RUN chmod +x /usr/bin/yaml2json

ARG VERSION_JD=1.4.0
ADD https://github.com/josephburnett/jd/releases/download/v${VERSION_JD}/jd-amd64-linux /usr/bin/jd
RUN chmod +x /usr/bin/jd

ARG VERSION_MILLER=5.10.2
ADD https://github.com/johnkerl/miller/releases/download/v${VERSION_MILLER}/mlr.linux.x86_64 /usr/bin/mlr
RUN chmod +x /usr/bin/mlr

ARG VERSION_GOMPLATE=3.9.0
ADD https://github.com/hairyhenderson/gomplate/releases/download/v${VERSION_GOMPLATE}/gomplate_linux-amd64 /usr/bin/gomplate
RUN chmod +x /usr/bin/gomplate

USER vscode

# Db2 (must run as vscode)
RUN go get -d github.com/ibmdb/go_ibm_db && \
    cd /home/vscode/go/pkg/mod/github.com/ibmdb/go_ibm_db@v0.4.1/installer && \
    go run setup.go

ENV DB2HOME=/home/vscode/go/pkg/mod/github.com/ibmdb/clidriver
ENV CGO_CFLAGS=-I$DB2HOME/include \
    CGO_LDFLAGS=-L$DB2HOME/lib \
    LD_LIBRARY_PATH=$DB2HOME/lib
