FROM alpine
RUN apk --update upgrade && \
    apk add ca-certificates && \
    apk add tzdata && \
    rm -rf /var/cache/apk/*
ADD frp-adapter /bin/
ADD .kube /root
CMD /bin/frp-adapter
