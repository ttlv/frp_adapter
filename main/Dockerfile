FROM ubuntu:latest
RUN mkdir -p /root/.kube
ADD frp-adapter /bin/
ADD config /root/.kube
CMD /bin/frp-adapter
