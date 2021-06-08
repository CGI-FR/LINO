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

FROM gcr.io/distroless/base

COPY bin/lino /usr/bin/lino

WORKDIR /home/lino

ENTRYPOINT [ "/usr/bin/lino" ]

ARG BUILD_DATE
ARG VERSION
ARG REVISION

# https://github.com/opencontainers/image-spec/blob/master/annotations.md
LABEL org.opencontainers.image.created       "${BUILD_DATE}"
LABEL org.opencontainers.image.authors       "CGI Lino <lino.fr@cgi.com>"
LABEL org.opencontainers.image.url           "https://github.com/CGI-FR/LINO"
LABEL org.opencontainers.image.documentation "https://github.com/CGI-FR/LINO/blob/main/README.md"
LABEL org.opencontainers.image.source        "https://github.com/CGI-FR/LINO.git"
LABEL org.opencontainers.image.version       "${VERSION}"
LABEL org.opencontainers.image.revision      "${REVISION}"
LABEL org.opencontainers.image.vendor        "CGI France"
LABEL org.opencontainers.image.licenses      "GPL-3.0-only"
LABEL org.opencontainers.image.ref.name      "cgi-lino"
LABEL org.opencontainers.image.title         "CGI LINO"
LABEL org.opencontainers.image.description   "LINO is a simple ETL (Extract Transform Load) tools to manage tests datas."
