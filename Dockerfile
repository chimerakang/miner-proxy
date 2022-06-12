FROM golang:1.17.6-alpine3.15

ADD . /home/miner-proxy

ENV GOPROXY=https://goproxy.io,direct

RUN cd /home/miner-proxy && go mod tidy && cd ./cmd/miner-proxy && go build .

WORKDIR /home
RUN cp /home/miner-proxy/config.json /home/

RUN mv /home/miner-proxy/docker-entrypoint /usr/bin/ && \
    mv /home/miner-proxy/cmd/miner-proxy/miner-proxy /usr/bin/ && \
    rm -rf /home/miner-proxy

EXPOSE 19999
EXPOSE 19998
RUN ls -l

ENTRYPOINT ["docker-entrypoint"]

CMD ["miner-proxy","-h"]
