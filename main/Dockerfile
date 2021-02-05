FROM alpine:latest
RUN apk --update upgrade && \
    apk add ca-certificates && \
    apk add tzdata && \
    rm -rf /var/cache/apk/*
RUN mkdir -p /root/.kube
ADD frp-adapter /bin/
ADD server.pem /root
ADD server.key /root
CMD /bin/frp-adapter
