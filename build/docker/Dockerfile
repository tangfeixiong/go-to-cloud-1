#FROM quay.io/coreos/flannelbox:1.0
#FROM alpine:3.4
#FROM gliderlabs/alpine
FROM openshift/origin-base
MAINTAINER tangfeixiong <fxtang@qingyuanos.com>

#ADD ./flanneld /opt/bin/
#ADD ./mk-docker-opts.sh /opt/bin/
#CMD /opt/bin/flanneld

#RUN apk add --update bash ca-certificates git libc6-compat && rm -rf /var/cache/apk/*
#RUN INSTALL_PKGS="which git tar wget hostname sysvinit-tools util-linux bsdtar epel-release \
#    socat ethtool device-mapper iptables tree findutils nmap-ncat e2fsprogs xfsprogs lsof" && \
#    yum install -y $INSTALL_PKGS && \
#    rpm -V $INSTALL_PKGS && \
#    yum clean all

#ADD apaas /bin/
ADD gotopaas /bin/apaas
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
ENV ETCD_V3_ADDRESSES 10.3.0.212:2379

VOLUME ["/root/.kube", "/openshift.local.config"]

EXPOSE 50051

CMD ["/bin/apaas"]

#ENTRYPOINT ["/bin/apaas"]

LABEL name="apaas" version="0.2" description="openshift origin, GitVersion: v1.3.0-alpha.2"
