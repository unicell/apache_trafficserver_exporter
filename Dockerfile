FROM       alpine:3.5
MAINTAINER Qiu Yu <unicell@gmail.com>
EXPOSE     9122

RUN true \
    && apk update \
    && apk --no-cache add ca-certificates curl \
    && rm -rf "/tmp/*" "/root/.cache" `find / -regex '.*\.py[co]'`

ADD  trafficserver_exporter /trafficserver_exporter
ADD  trafficserver-mapping.conf /trafficserver-mapping.conf

CMD ["/trafficserver_exporter"]
