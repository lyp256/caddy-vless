FROM docker.io/library/golang:1 as build
COPY . /src
RUN cd /src && CGO_ENABLED=0 make build
RUN chmod +x /src/build/proxy

FROM docker.io/library/alpine:3 as prod
COPY --from=build /src/build/* /usr/local/bin/
CMD proxy
