FROM alpine:latest
RUN apk --update upgrade && \
    apk add ca-certificates && \
    apk add tzdata && \
    rm -rf /var/cache/apk/*
RUN mkdir -p /root/.kube
ADD frp-adapter /bin/
ADD config.ini /bin/
CMD /bin/frp-adapter
