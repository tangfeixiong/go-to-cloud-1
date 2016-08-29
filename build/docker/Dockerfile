#FROM quay.io/coreos/flannelbox:1.0
#FROM alpine:3.4
FROM gliderlabs/alpine
MAINTAINER tangfeixiong <fxtang@qingyuanos.com>

#ADD ./flanneld /opt/bin/
#ADD ./mk-docker-opts.sh /opt/bin/
#CMD /opt/bin/flanneld

RUN apk add --update bash ca-certificates git libc6-compat && rm -rf /var/cache/apk/*

ADD apaas /bin/
ADD ./openshift.local.config/ /openshift.local.config/
ADD ./ssl/ /root/.kube/
ADD ./openshift.local.config/kubeconfig /root/.kube/config
ADD ./openshift.local.config/admin.kubeconfig /openshift.local.config/master/admin.kubeconfig

ENV KUBE_CONFIG /root/.kube/config
# ENV KUBE_CONTEXT kube
ENV KUBE_CONTEXT openshift-origin-single
ENV OSO_CONFIG /openshift.local.config/master/admin.kubeconfig
# ENV OSO_CONTEXT default/20-0-0-64:8443/system:admin
ENV OSO_CONTEXT tangfx/172-17-4-50:30443/system:admin
# ENV ORIGIN_VERSION v1.3.0-alpha.3
ENV ORIGIN_VERSION v1.3.0-alpha.1
ENV APAAS_GRPC_PORT :50051
ENV GNATSD_ADDRESSES 10.3.0.39:4222

VOLUME ["/root/.kube", "/openshift.local.config"]

EXPOSE 50051

CMD ["/bin/apaas"]

#ENTRYPOINT ["/bin/apaas"]

LABEL name="apaas-build" version="0.1" description="openshift origin, GitVersion: v1.3.0-alpha.2"
