FROM quay.io/coreos/flannelbox:1.0
MAINTAINER tangfeixiong <fxtang@qingyuanos.com>

#ADD ./flanneld /opt/bin/
#ADD ./mk-docker-opts.sh /opt/bin/
#CMD /opt/bin/flanneld

LABEL name="apaas-build" version="0.1" description="openshift origin, GitVersion: v1.3.0-alpha.2"

ADD apaas /bin/
ADD kubeconfig /root/.kube/config
#ADD ./ssl /root/ssl

ENV PORT 50051
EXPOSE 50051

CMD ["/bin/apaas"]

#ENTRYPOINT ["/bin/apaas"]
