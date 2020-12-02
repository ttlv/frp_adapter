FROM ubuntu:latest
RUN mkdir -p /root/.kube
ADD frp-adapter /bin/
ADD config /root/.kube
ADD cert.pem /root
ADD key.pem /root
CMD /bin/frp-adapter
