FROM ubuntu:latest
RUN mkdir -p /root/.kube
ADD frp-adapter /bin/
ADD config /root/.kube
ADD server.pem /root
ADD server.key /root
CMD /bin/frp-adapter
