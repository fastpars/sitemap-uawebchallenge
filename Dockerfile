FROM fastpars/golnag-debian-env:latest

MAINTAINER Mykhailo Tsiuptsiun <fastpars@gmail.com>

COPY gopath /gopath
COPY www /www

RUN go install sitemap

EXPOSE 8888
ENTRYPOINT /gopath/bin/sitemap
