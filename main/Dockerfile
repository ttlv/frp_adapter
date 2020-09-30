FROM alpine
RUN apk --update upgrade && \
    apk add ca-certificates && \
    apk add tzdata && \
    rm -rf /var/cache/apk/*

ADD frp-adapter /bin/
CMD /bin/frp-adapter
