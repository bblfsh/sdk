FROM ${DOCKER_BUILD_NATIVE_IMAGE}

ENV GOLANG_SRC_URL https://golang.org/dl/go${RUNTIME_GO_VERSION}.src.tar.gz

# from https://github.com/docker-library/golang/blob/master/1.8/alpine/Dockerfile

RUN apk add --update --no-cache ca-certificates openssl && update-ca-certificates
RUN wget https://raw.githubusercontent.com/docker-library/golang/132cd70768e3bc269902e4c7b579203f66dc9f64/1.8/alpine/no-pic.patch

RUN set -ex \
	&& apk add --no-cache --virtual .build-deps \
		bash \
		gcc \
		musl-dev \
		openssl \
		go \
	\
	&& export GOROOT_BOOTSTRAP="$(go env GOROOT)" \
	\
	&& wget -q "$GOLANG_SRC_URL" -O golang.tar.gz \
	&& tar -C /usr/local -xzf golang.tar.gz \
	&& rm golang.tar.gz \
	&& cd /usr/local/go/src \
	&& patch -p2 -i /no-pic.patch \
	&& ./make.bash \
	\
	&& rm -rf /*.patch \
	&& apk del .build-deps

ENV GOPATH /go
ENV PATH $GOPATH/bin:/usr/local/go/bin:$PATH

RUN mkdir -p "$GOPATH/src" "$GOPATH/bin" && chmod -R 777 "$GOPATH"