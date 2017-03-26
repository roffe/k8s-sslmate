FROM debian:jessie-slim

RUN apt-get update && apt-get install -y wget

RUN wget -P /etc/apt/sources.list.d https://sslmate.com/apt/jessie/sslmate.list
RUN wget -P /etc/apt/trusted.gpg.d https://sslmate.com/apt/jessie/sslmate.gpg
RUN apt-get update \
    && apt-get install -y \
    sslmate \
    && rm -rf /var/lib/apt/lists/*

COPY scripts/* /

RUN chmod a+x /start.sh

ENTRYPOINT ["/start.sh"]