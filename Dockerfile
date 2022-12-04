# build the web UI
FROM node:16-alpine as frontend

ADD . /build
WORKDIR /build/frontend

RUN yarn install && yarn build

# build the backend registry-admin
FROM golang:1.19-alpine as backend

ARG GIT_BRANCH
ARG GITHUB_SHA
ARG CI

ENV GOFLAGS="-mod=vendor"
ENV CGO_ENABLED=1

ADD . /build
COPY --from=frontend /build/frontend/build  /build/app/web

RUN apk add --no-cache --update git tzdata ca-certificates build-base git

WORKDIR /build

RUN \
    if [ -z "$CI" ] ; then \
    echo "runs outside of CI" && version=$(git rev-parse --abbrev-ref HEAD)-$(git log -1 --format=%h)-$(date +%Y%m%dT%H:%M:%S); \
    else version=${GIT_BRANCH}-${GITHUB_SHA:0:7}-$(date +%Y%m%dT%H:%M:%S); fi && \
    echo "version=$version" && \
    cd app && go mod vendor &&  go build -o /build/registry-admin -ldflags "-X main.version=${version} -s -w"


FROM umputun/baseimage:app-v1.9.2

LABEL org.opencontainers.image.authors="zebox <zebox@ya.ru>" \
      org.opencontainers.image.description="Registry administartion UI tool" \
      org.opencontainers.image.licenses="MIT" \
      org.opencontainers.image.source="https://github.com/zebox/registry-admin.git" \
      org.opencontainers.image.title="Registry Admin" \
      org.opencontainers.image.url="https://github.com/zebox/registry-admin"

COPY --from=backend /build/registry-admin /app/registry-admin

WORKDIR /app

RUN chown -R app:app /app
RUN ln -s /app/registry-admin /usr/bin/registry-admin

EXPOSE 80
EXPOSE 443

CMD ["/app/registry-admin"]