FROM alpine:latest
RUN mkdir -p /root/.kube
ADD frp-adapter /bin/
ADD server.pem /root
ADD server.key /root
CMD /bin/frp-adapter
