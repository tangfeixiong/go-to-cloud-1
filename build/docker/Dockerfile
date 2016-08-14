FROM quay.io/coreos/flannelbox:1.0
MAINTAINER tangfeixiong <fxtang@qingyuanos.com>

#ADD ./flanneld /opt/bin/
#ADD ./mk-docker-opts.sh /opt/bin/
#CMD /opt/bin/flanneld

LABEL name="apaas-build" version="0.1" description="openshift origin, GitVersion: v1.3.0-alpha.2"

ADD apaas /bin/
#ADD ./openshift.local.config/kubeconfig /root/.kube/config
ADD ./ssl/kubeconfig /root/.kube/config
ADD ./ssl/ /root/.kube/
ADD ./openshift.local.config/ /openshift.local.config/

ENV PORT 50051
ENV K8S_KUBECONFIG /root/.kube/config
ENV K8S_KUBECONTEXT kube
ENV OSO_KUBECONFIG /openshift.local.config/master/admin.kubeconfig
ENV OSO_KUBECONTEXT default/127-0-0-1:8443/system:admin

VOLUME ["/root/.kube", "/openshift.local.config"]

EXPOSE 50051

CMD ["/bin/apaas"]

#ENTRYPOINT ["/bin/apaas"]
