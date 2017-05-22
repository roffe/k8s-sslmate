FROM debian:jessie-slim

RUN apt-get update && apt-get install -y wget unzip

RUN wget -P /etc/apt/sources.list.d https://sslmate.com/apt/jessie/sslmate.list
RUN wget -P /etc/apt/trusted.gpg.d https://sslmate.com/apt/jessie/sslmate.gpg
RUN apt-get update \
    && apt-get install -y \
    sslmate \
    && rm -rf /var/lib/apt/lists/*

COPY scripts/* /

RUN mkdir -p /opt/bin
RUN chmod a+x /start.sh

ADD https://github.com/tianon/gosu/releases/download/1.10/gosu-amd64 /opt/bin/gosu
RUN chmod a+x /opt/bin/gosu

ADD https://roffe.nu/k8s-sslmate/k8s-sslmate.zip /root
RUN unzip /root/k8s-sslmate.zip -d /opt/bin/ \
	&& chmod a+x /opt/bin/k8s-sslmate \
	&& rm -rf /root/k8s-sslmate.zip
#COPY bin/k8s-sslmate /opt/bin/
RUN chmod a+x /opt/bin/k8s-sslmate

ENTRYPOINT ["/start.sh"]