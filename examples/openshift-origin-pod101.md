# Openshift Origin hands on - VirtualBox

MacBook-Pro:trusty64 fanhongling$ VBoxManage --version
5.0.20r106931

git clone https://github.com/openshift/origin.git origin

vagran ssh

[vagrant@localhost ~]$ sudo password vagrant
sudo: password：找不到命令
[vagrant@localhost ~]$ sudo passwd vagrant

[vagrant@localhost ~]$ vim .bashrc 
[vagrant@localhost ~]$ export PATH=/data/bin:$PATH

[vagrant@localhost origin]$ make build GOFLAGS=-v


[vagrant@localhost origin]$ cp _output/local/bin/linux/amd64/openshift /data/bin

[vagrant@localhost origin]$ openshift version
openshift v1.3.0-alpha.0-110-g3db0fc4
kubernetes v1.3.0-alpha.1-331-g0522e63
etcd 2.3.0


[vagrant@localhost origin]$ cp _output/local/bin/linux/amd64/oc /data/bin

[vagrant@localhost origin]$ oc version
oc v1.3.0-alpha.0-110-g3db0fc4
kubernetes v1.3.0-alpha.1-331-g0522e63

[vagrant@localhost origin]$ mkdir openshift.local.config/master/

[vagrant@localhost origin]$ cp etc/kubernetes/kubeconfig openshift.local.config/master/


[vagrant@localhost origin]$ openshift --kubeconfig=openshift.local.config/master/kubeconfig kube get ns
NAME          STATUS    AGE
default       Active    22m
kube-system   Active    20m

[vagrant@localhost origin]$ oc --config=openshift.local.config/master/kubeconfig get nodes
NAME          STATUS    AGE
172.17.4.50   Ready     23m


## kubernetes V1.3.0-alpha.2


[vagrant@localhost origin]$ ./build-docker-image.sh
[vagrant@localhost origin]$ docker images
REPOSITORY                                       TAG                  IMAGE ID            CREATED             VIRTUAL SIZE
tangfeixiong/openshift-origin                    latest               e21f3ead5f3c        3 hours ago         473.8 MB
docker.io/openshift/origin                       latest               d0d0bf52d244        8 hours ago         415.5 MB
gcr.io/google_containers/hyperkube-amd64         v1.2.4               b65f775dbf89        4 days ago          316.7 MB
gcr.io/google_containers/hyperkube-amd64         v1.3.0-alpha.2       c11527c21c02        4 weeks ago         398.5 MB
gcr.io/google_containers/exechealthz             1.0                  d6fccb55b399        5 weeks ago         7.116 MB
quay.io/tangfeixiong/netcat-http-server-simple   latest               7fa32f504c61        7 weeks ago         7.807 MB
gcr.io/google_containers/kube2sky                1.14                 c0b611ff3f70        9 weeks ago         27.8 MB
openshift/origin-release                         latest               6791072422b1        9 weeks ago         715.2 MB
openshift/origin-haproxy-router-base             latest               a0328f433acf        9 weeks ago         290.9 MB
openshift/origin-base                            latest               f2ffca9a8520        9 weeks ago         271.8 MB
docker.io/centos                                 centos7              2933d50b9f77        11 weeks ago        196.6 MB
gcr.io/google_containers/etcd-amd64              2.2.1                202873aab189        3 months ago        28.19 MB
gcr.io/google_containers/skydns                  2015-10-13-8c72f8c   d8ed451aa9b9        7 months ago        40.55 MB
gcr.io/google_containers/pause                   2.0                  8950680a606c        7 months ago        350.2 kB
docker.io/openshift/etcd-20-centos7              latest               7857141e9bb1        10 months ago       244.3 MB



* Create Openshift Origin namespace and service

[vagrant@localhost origin]$ oc --config=openshift.local.config/master/kubeconfig create -f openshift-origin-namespace.yaml 
namespace "openshift-origin" created

[vagrant@localhost origin]$ kubectl --kubeconfig=etc/kubernetes/kubeconfig get ns
NAME               STATUS    AGE
default            Active    29m
kube-system        Active    26m
openshift-origin   Active    14s


[vagrant@localhost origin]$ oc --config=openshift.local.config/master/kubeconfig create -f openshift-service.yaml --namespace=openshift-origin
service "openshift" created

[vagrant@localhost origin]$ oc --config=openshift.local.config/master/kubeconfig --namespace=openshift-origin get svc
NAME        CLUSTER-IP   EXTERNAL-IP   PORT(S)    AGE
openshift   10.3.0.179                 8443/TCP   29s


[vagrant@localhost origin]$ oc --config=openshift.local.config/master/kubeconfig --namespace=openshift-origin get svc openshift -o jsonpath={.spec.ports[0].nodePort};echo
30448

* Create Openshift Origin data store - etcd

[vagrant@localhost origin]$ openshift --config=openshift.local.config/master/kubeconfig kube create -f etcd-standalone-pv.yaml 
persistentvolume "etcd-storage" created

[vagrant@localhost origin]$ openshift --kubeconfig=openshift.local.config/master/kubeconfig kube get pv
NAME           CAPACITY   ACCESSMODES   STATUS      CLAIM     REASON    AGE
etcd-storage   5Gi        RWO,ROX,RWX   Available                       47s

[vagrant@localhost origin]$ oc --config=openshift.local.config/master/kubeconfig create -f etcd-standalone-pvc.yaml 
persistentvolumeclaim "etcd-storage" created

[vagrant@localhost origin]$ oc --config=openshift.local.config/master/kubeconfig --namespace=openshift-origin get pvc
NAME           STATUS    VOLUME         CAPACITY   ACCESSMODES   AGE
etcd-storage   Bound     etcd-storage   5Gi        RWO,ROX,RWX   25s

[vagrant@localhost origin]$ oc --config=openshift.local.config/master/kubeconfig create -f etcd-standalone-service.json 
service "etcd" created

[vagrant@localhost origin]$ oc --config=openshift.local.config/master/kubeconfig --namespace=openshift-origin get svc
NAME        CLUSTER-IP   EXTERNAL-IP   PORT(S)             AGE
etcd        10.3.0.16    <none>        4001/TCP,7001/TCP   9s
openshift   10.3.0.179                 8443/TCP            4m

[vagrant@localhost origin]$ oc --config=openshift.local.config/master/kubeconfig create -f etcd-standalone-controller.json
replicationcontroller "etcd" created

[vagrant@localhost origin]$ oc --config=openshift.local.config/master/kubeconfig --namespace=openshift-origin get pods
NAME         READY     STATUS    RESTARTS   AGE
etcd-xxqt2   1/1       Running   0          13s

* Create Openshift Origin master configuration


[vagrant@localhost origin]$ openshift start master --kubeconfig=openshift.local.config/master/kubeconfig --write-config=openshift.local.config/master --dns=tcp://10.3.0.10:53 --host-subnet-length=7 --network-cidr=172.17.0.1/22 --portal-net=10.3.0.0/24 --etcd=http://10.3.0.16:4001 --master=https://10.3.0.179:8443 --public-master=https://172.17.4.50:30448
Wrote master config to: openshift.local.config/master/master-config.yaml

[vagrant@localhost origin]$ cp etc/kubernetes/cacerts/server.key openshift.local.config/master/k8s-apiserver.key

[vagrant@localhost origin]$ oc --config=openshift.local.config/master/kubeconfig --namespace=openshift-origin get serviceaccounts
NAME      SECRETS   AGE
default   1         14m

[vagrant@localhost origin]$ oc --config=openshift.local.config/master/kubeconfig --namespace=openshift-origin get secrets
NAME                  TYPE                                  DATA      AGE
default-token-hxgp1   kubernetes.io/service-account-token   3         14m

[vagrant@localhost origin]$ cp openshift-controller.yaml openshift-controller-temporay.yaml 
[vagrant@localhost origin]$ vim openshift-controller-temporay.yaml 

[vagrant@localhost origin]$ oc --config=openshift.local.config/master/kubeconfig --namespace=openshift-origin create -f openshift-controller-temporay.yaml 
replicationcontroller "openshift" created

[vagrant@localhost origin]$ docker logs 7bf80324f120
W0511 01:10:29.603779       1 start_master.go:270] assetConfig.loggingPublicURL: Invalid value: "": required to view aggregated container logs in the console
W0511 01:10:29.615353       1 start_master.go:270] assetConfig.metricsPublicURL: Invalid value: "": required to view cluster metrics in the console
I0511 01:10:29.649550       1 admission.go:36] Admission plugin "ProjectRequestLimit" is not configured so it will be disabled.
I0511 01:10:29.650172       1 admission.go:33] Admission plugin "PodNodeConstraints" is not configured so it will be disabled.
I0511 01:10:29.667372       1 start_master.go:383] Starting master on 0.0.0.0:8443 (v1.3.0-alpha.0-443-g225bc84)
I0511 01:10:29.667834       1 start_master.go:384] Public master address is https://172.17.4.50:30448
I0511 01:10:29.670037       1 start_master.go:388] Using images from "openshift/origin-<component>:v1.3.0-alpha.0"
I0511 01:10:29.686812       1 run_components.go:205] Using default project node label selector: 
W0511 01:10:29.778207       1 swagger.go:32] No API exists for predefined swagger description /api/v1
I0511 01:10:29.778961       1 master.go:264] Started Kubernetes proxy at 0.0.0.0:8443/api/
I0511 01:10:29.779385       1 master.go:264] Started Origin API at 0.0.0.0:8443/oapi/v1
I0511 01:10:29.779628       1 master.go:264] Started OAuth2 API at 0.0.0.0:8443/oauth
I0511 01:10:29.779852       1 master.go:264] Started Web Console 0.0.0.0:8443/console/
I0511 01:10:29.780071       1 master.go:264] Started Swagger Schema API at 0.0.0.0:8443/swaggerapi/
...

[vagrant@localhost origin]$ kubectl --kubeconfig=openshift.local.config/master/kubeconfig --namespace=openshift-origin get serviceaccounts
NAME       SECRETS   AGE
builder    2         58s
default    2         1h
deployer   2         58s

[vagrant@localhost origin]$ kubectl --kubeconfig=openshift.local.config/master/kubeconfig --namespace=openshift-origin get secrets
NAME                       TYPE                                  DATA      AGE
builder-dockercfg-9u99p    kubernetes.io/dockercfg               1         2m
builder-token-3ns66        kubernetes.io/service-account-token   3         2m
builder-token-zwepq        kubernetes.io/service-account-token   3         2m
default-dockercfg-ina7d    kubernetes.io/dockercfg               1         2m
default-token-evikg        kubernetes.io/service-account-token   3         2m
default-token-hxgp1        kubernetes.io/service-account-token   3         1h
deployer-dockercfg-10545   kubernetes.io/dockercfg               1         2m
deployer-token-dxex0       kubernetes.io/service-account-token   3         2m
deployer-token-iukrh       kubernetes.io/service-account-token   3         2m

[vagrant@localhost origin]$ oc --config=openshift.local.config/master/admin.kubeconfig secrets new openshift-config openshift.local.config/master --output=json | tee secret.json
{
    "kind": "Secret",
    "apiVersion": "v1",
    "metadata": {
        "name": "openshift-config",
        "creationTimestamp": null
    },
    "data": {
        "admin.crt": "
...
"
    },
    "type": "Opaque"
}

[vagrant@localhost origin]$ oc --config=openshift.local.config/master/admin.kubeconfig --namespace=openshift-origin create -f secret.json 
secret "openshift-config" created

[vagrant@localhost origin]$ oc --config=openshift.local.config/master/admin.kubeconfig --namespace=openshift-origin secrets add serviceaccounts/builder secrets/openshift-config

[vagrant@localhost origin]$ oc --config=openshift.local.config/master/admin.kubeconfig --namespace=openshift-origin get secrets
NAME                       TYPE                                  DATA      AGE
builder-dockercfg-9u99p    kubernetes.io/dockercfg               1         7m
builder-token-3ns66        kubernetes.io/service-account-token   3         7m
builder-token-zwepq        kubernetes.io/service-account-token   3         7m
default-dockercfg-ina7d    kubernetes.io/dockercfg               1         7m
default-token-evikg        kubernetes.io/service-account-token   3         7m
default-token-hxgp1        kubernetes.io/service-account-token   3         1h
deployer-dockercfg-10545   kubernetes.io/dockercfg               1         7m
deployer-token-dxex0       kubernetes.io/service-account-token   3         7m
deployer-token-iukrh       kubernetes.io/service-account-token   3         7m
openshift-config           Opaque                                31        32s

[vagrant@localhost origin]$ oc --config=openshift.local.config/master/admin.kubeconfig get sa --all-namespaces=true
NAMESPACE          NAME                               SECRETS   AGE
default            builder                            2         9m
default            default                            2         2h
default            deployer                           2         9m
kube-system        builder                            2         9m
kube-system        default                            2         2h
kube-system        deployer                           2         9m
openshift          builder                            2         9m
openshift          default                            2         9m
openshift          deployer                           2         9m
openshift-infra    build-controller                   2         9m
openshift-infra    builder                            2         9m
openshift-infra    daemonset-controller               2         9m
openshift-infra    default                            2         9m
openshift-infra    deployer                           2         9m
openshift-infra    deployment-controller              2         9m
openshift-infra    gc-controller                      2         9m
openshift-infra    hpa-controller                     2         9m
openshift-infra    job-controller                     2         9m
openshift-infra    namespace-controller               2         9m
openshift-infra    pv-binder-controller               2         9m
openshift-infra    pv-provisioner-controller          2         9m
openshift-infra    pv-recycler-controller             2         9m
openshift-infra    replication-controller             2         9m
openshift-infra    service-load-balancer-controller   2         9m
openshift-origin   builder                            3         9m
openshift-origin   default                            2         1h
openshift-origin   deployer                           2         9m

[vagrant@localhost origin]$ docker login qingyuanos.com
Username: ***
Password: 
Email: fxtang@qingyuanos.com
WARNING: login credentials saved in /home/vagrant/.docker/config.json
Login Succeeded

[vagrant@localhost origin]$ oc --config=openshift.local.config/master/admin.kubeconfig --namespace=openshift-origin secrets newtangfeixiong .dockerconfigjson=/home/vagrant/.docker/config.json
secret/tangfeixiong




[vagrant@localhost origin]$ oc --config=openshift.local.config/master/admin.kubeconfig login
Authentication required for https://172.17.4.50:30448 (openshift)
Username: tangfeixiong
Password: 
Login successful.

You don't have any projects. You can try to create a new project, by running

    $ oc new-project <projectname>

[vagrant@localhost origin]$ oc --config=openshift.local.config/master/admin.kubeconfig new-project tangfeixiong
Now using project "tangfeixiong" on server "https://172.17.4.50:30448".

You can add applications to this project with the 'new-app' command. For example, try:

    $ oc new-app centos/ruby-22-centos7~https://github.com/openshift/ruby-ex.git

to build a new example application in Ruby.


[vagrant@localhost origin]$ oc --config=openshift.local.config/master/admin.kubeconfig --namespace=tangfeixiong new-build https://github.com/tangfeixiong/docker-nc.git#master --context-dir=latest --strategy=docker --name=nc --to-docker=true  --allow-missing-images=true --output=json
{
    "kind": "List",
    "apiVersion": "v1",
    "metadata": {},
    "items": [
        {
            "kind": "ImageStream",
            "apiVersion": "v1",
            "metadata": {
                "name": "alpine",
                "creationTimestamp": null,
                "labels": {
                    "build": "nc"
                },
                "annotations": {
                    "openshift.io/generated-by": "OpenShiftNewBuild"
                }
            },
            "spec": {
                "tags": [
                    {
                        "name": "latest",
                        "annotations": {
                            "openshift.io/imported-from": "alpine:latest"
                        },
                        "from": {
                            "kind": "DockerImage",
                            "name": "alpine:latest"
                        },
                        "generation": null,
                        "importPolicy": {}
                    }
                ]
            },
            "status": {
                "dockerImageRepository": ""
            }
        },
        {
            "kind": "BuildConfig",
            "apiVersion": "v1",
            "metadata": {
                "name": "nc",
                "creationTimestamp": null,
                "labels": {
                    "build": "nc"
                },
                "annotations": {
                    "openshift.io/generated-by": "OpenShiftNewBuild"
                }
            },
            "spec": {
                "triggers": [
                    {
                        "type": "GitHub",
                        "github": {
                            "secret": "5Z7DPeg3Ul2rJIAGtzJ7"
                        }
                    },
                    {
                        "type": "Generic",
                        "generic": {
                            "secret": "BJS3bVeDxKjg1Zl4RZWY"
                        }
                    },
                    {
                        "type": "ConfigChange"
                    },
                    {
                        "type": "ImageChange",
                        "imageChange": {}
                    }
                ],
                "source": {
                    "type": "Git",
                    "git": {
                        "uri": "https://github.com/tangfeixiong/docker-nc.git",
                        "ref": "master"
                    },
                    "contextDir": "latest",
                    "secrets": []
                },
                "strategy": {
                    "type": "Docker",
                    "dockerStrategy": {
                        "from": {
                            "kind": "ImageStreamTag",
                            "name": "alpine:latest"
                        }
                    }
                },
                "output": {
                    "to": {
                        "kind": "DockerImage",
                        "name": "nc:latest"
                    },
					 "pushSecret": {
                        "name": "qingyuanos"
					}
                },
                "resources": {},
                "postCommit": {}
            },
            "status": {
                "lastVersion": 0
            }
        }
    ]
}

















[vagrant@localhost ~]$ vim openshift-controller.yaml 

[vagrant@localhost ~]$ kubectl --kubeconfig=coreos.kubeconfig create -f openshift-controller.yaml 
replicationcontroller "openshift" created

[vagrant@localhost origin]$ oc --config=openshift.local.config/master/kubeconfig --namespace=openshift-origin create -f openshift-controller.yaml 
replicationcontroller "openshift" created

[vagrant@localhost origin]$ oc --config=openshift.local.config/master/admin.kubeconfig config view
apiVersion: v1
clusters:
- cluster:
    certificate-authority-data: REDACTED
    server: https://10.3.0.179:8443
  name: 10-3-0-179:8443
- cluster:
    certificate-authority-data: REDACTED
    server: https://172.17.4.50:30448
  name: 172-17-4-50:30448
contexts:
- context:
    cluster: 10-3-0-179:8443
    namespace: default
    user: system:admin/10-3-0-179:8443
  name: default/10-3-0-179:8443/system:admin
- context:
    cluster: 172-17-4-50:30448
    namespace: default
    user: system:admin/10-3-0-179:8443
  name: default/172-17-4-50:30448/system:admin
current-context: default/10-3-0-179:8443/system:admin
kind: Config
preferences: {}
users:
- name: system:admin/10-3-0-179:8443
  user:
    client-certificate-data: REDACTED
    client-key-data: REDACTED

[vagrant@localhost origin]$ oc --config=openshift.local.config/master/admin.kubeconfig config use-context default/172-17-4-50:30448/system:admin
switched to context "default/172-17-4-50:30448/system:admin".

[centos@k8s-master-ha-2 amd64]$ ./oc --config=/home/centos/.kube/config secrets new openshift-local-config /home/centos/devops/amd64/openshift.local.config/master -o json | tee secret.json
error: the server could not find the requested resource
See 'oc secrets new -h' for help and examples.


* from CoreOS single

fanhonglingdeMacBook-Pro:single-node fanhongling$ scp -r kubeconfig ssl/ vagrant@172.17.4.50:/data/src/github.com/openshift/origin
vagrant@172.17.4.50's password: 
kubeconfig                                  100%  437     0.4KB/s   00:00    
admin-key.pem                               100% 1675     1.6KB/s   00:00    
admin-req.cnf                               100%  290     0.3KB/s   00:00    
admin.csr                                   100% 1005     1.0KB/s   00:00    
admin.pem                                   100% 1078     1.1KB/s   00:00    
apiserver-key.pem                           100% 1679     1.6KB/s   00:00    
apiserver-req.cnf                           100%  320     0.3KB/s   00:00    
apiserver.csr                               100% 1021     1.0KB/s   00:00    
apiserver.pem                               100% 1094     1.1KB/s   00:00    
ca-key.pem                                  100% 1679     1.6KB/s   00:00    
ca.pem                                      100% 1135     1.1KB/s   00:00    
ca.srl                                      100%   17     0.0KB/s   00:00    
controller.tar                              100% 7680     7.5KB/s   00:00    
kube-admin.tar                              100% 7680     7.5KB/s   00:00    

* back Openshift fedora inst

[vagrant@localhost ~]$

[vagrant@localhost ~]$ kubectl --kubeconfig=kubeconfig get nodes
NAME          STATUS    AGE
172.17.4.99   Ready     96d


[vagrant@localhost ~]$ cp Godeps/_workspace/src/k8s.io/kubernetes/examples/openshift-origin/openshift-*.yaml .

* Create openshift namespace

grant@localhost origin]$ kubectl --kubeconfig=kubeconfig create -f openshift-origin-namespace.yaml 
namespace "openshift-infra" created

[vagrant@localhost ~]$ kubectl --kubeconfig=kubeconfig get namespaces
NAME               STATUS    AGE
default            Active    97d
helm               Active    1d
kube-system        Active    97d
openshift-origin   Active    9s

[vagrant@localhost origin]$ kubectl --kubeconfig=kubeconfig create -f etcd-standalone-service.json 
service "etcd" created

fanhonglingdeMacBook-Pro:single-node fanhongling$ kubectl --kubeconfig=kubeconfig --namespace=openshift-origin get svc
NAME        CLUSTER-IP   EXTERNAL-IP   PORT(S)             AGE
etcd        10.3.0.32    <none>        4001/TCP,7001/TCP   12s
openshift   10.3.0.216   nodes         8443/TCP            38m


* Change service type into NodePort

[vagrant@localhost ~]$ vi openshift-service.yaml 

[vagrant@localhost ~]$ kubectl --kubeconfig=kubeconfig create -f openshift-service.yaml 
You have exposed your service on an external port on all nodes in your
cluster.  If you want to expose this service to the external internet, you may
need to set up firewall rules for the service port(s) (tcp:32670) to serve traffic.

See http://releases.k8s.io/release-1.2/docs/user-guide/services-firewalls.md for more details.
service "openshift" created


[vagrant@localhost ~]$ kubectl --kubeconfig=kubeconfig --namespace=openshift-origin get svc
NAME        CLUSTER-IP   EXTERNAL-IP   PORT(S)    AGE
openshift   10.3.0.216   nodes         8443/TCP   1m


* Create openshift configuration

openshift start master --kubeconfig=openshift.local.config/master/kubeconfig --write-config=openshift.local.config/master --master=https://127.0.0.1:8443 --public-master=https://172.17.4.99:32670 --etcd=http://10.3.0.32:4001
Generated new key pair as openshift.local.config/master/serviceaccounts.public.key and openshift.local.config/master/serviceaccounts.private.key
Wrote master config to: openshift.local.config/master/master-config.yaml

[vagrant@localhost origin]$ cd ../../tangfeixiong/go-to-cloud-1/
[vagrant@localhost go-to-cloud-1]$ sudo /data/bin/openshift start --master-config=openshift.local.config/master/master-config.yaml --node-config=openshift.local.config/node-localhost/node-config.yaml --loglevel=2
W0505 19:47:39.264468   19437 start_master.go:270] assetConfig.loggingPublicURL: Invalid value: "": required to view aggregated container logs in the console
W0505 19:47:39.264942   19437 start_master.go:270] assetConfig.metricsPublicURL: Invalid value: "": required to view cluster metrics in the console
I0505 19:47:39.282963   19437 plugins.go:71] No cloud provider specified.

...

I0505 19:47:57.530219   19437 kubelet.go:2770] Recording NodeReady event message for node localhost
^Z
[1]+  已停止               sudo /data/bin/openshift start --master-config=openshift.local.config/master/master-config.yaml --node-config=openshift.local.config/node-localhost/node-config.yaml --loglevel=2
[vagrant@localhost go-to-cloud-1]$ bg %1
[1]+ sudo /data/bin/openshift start --master-config=openshift.local.config/master/master-config.yaml --node-config=openshift.local.config/node-localhost/node-config.yaml --loglevel=2 &


[vagrant@localhost go-to-cloud-1]$ cd ../../openshift/origin/
[vagrant@localhost go-to-cloud-1]$ oc secret new openshift-config openshift.local.config/master --output=json --config=/data/src/github.com/tangfeixiong/go-to-cloud-1/openshift.local.config/master/admin.kubeconfig | tee secret.json
{
    "kind": "Secret",
    "apiVersion": "v1",
    "metadata": {
        "name": "openshift-coreos-config",
        "creationTimestamp": null
    },
    "data": {
        "admin.crt": "...base64 encoding",
        "admin.key": "...base64 encoding",
        "admin.kubeconfig": "...base64 encoding",
        "ca-bundle.crt": "...base64 encoding",
        "ca.crt": "...base64 encoding",
        "ca.key": "...base64 encoding",
        "ca.serial.txt": "MEE=",
        "etcd.server.crt": "...base64 encoding",
        "etcd.server.key": "...base64 encoding",
        "master-config.yaml": "...base64 encoding",
        "master.etcd-client.crt": "...base64 encoding",
        "master.etcd-client.key": "...base64 encoding",
        "master.kubelet-client.crt": "...base64 encoding",
        "master.kubelet-client.key": "...base64 encoding",
        "master.proxy-client.crt": "...base64 encoding",
        "master.proxy-client.key": "...base64 encoding",
        "master.server.crt": "...base64 encoding",
        "master.server.key": "...base64 encoding",
        "openshift-master.crt": "...base64 encoding",
        "openshift-master.key": "...base64 encoding",
        "openshift-master.kubeconfig": "...base64 encoding",
        "openshift-registry.crt": "...base64 encoding",
        "openshift-registry.key": "...base64 encoding",
        "openshift-registry.kubeconfig": "...base64 encoding",
        "openshift-router.crt": "...base64 encoding",
        "openshift-router.key": "...base64 encoding",
        "openshift-router.kubeconfig": "...base64 encoding",
        "policy.json": "...base64 encoding",
        "serviceaccounts.private.key": "...base64 encoding",
        "serviceaccounts.public.key": "...base64 encoding"
    },
    "type": "Opaque"
}

* Change namespace into openshift

[vagrant@localhost ~]$ vim secret.json 

[vagrant@localhost ~]$ kubectl --kubeconfig=kubeconfig create -f secret.json 
secret "openshift-config" created

[vagrant@localhost origin]$ kubectl --kubeconfig=kubeconfig --namespace=openshift-origin get secrets
NAME                  TYPE                                  DATA      AGE
default-token-n8k2q   kubernetes.io/service-account-token   2         42m
openshift-config      Opaque                                30        17s



[vagrant@localhost origin]$ kubectl --kubeconfig=kubeconfig --namespace=openshift-origin get rc
NAME        DESIRED   CURRENT   AGE
openshift   1         1         41s
[vagrant@localhost origin]$ kubectl --kubeconfig=kubeconfig --namespace=openshift-origin get pods
NAME              READY     STATUS    RESTARTS   AGE
openshift-l573a   0/1       Pending   0          43s
[vagrant@localhost origin]$ kubectl --kubeconfig=kubeconfig --namespace=openshift-origin describe pods -l name=openshift
...
Volumes:
  config:
    Type:	Secret (a volume populated by a Secret)
    SecretName:	openshift-config
  default-token-n8k2q:
    Type:	Secret (a volume populated by a Secret)
    SecretName:	default-token-n8k2q
Events:
  FirstSeen	LastSeen	Count	From			SubobjectPath				Type		Reason		Message
  ---------	--------	-----	----			-------------				--------	------		-------
  2m		2m		1	{scheduler }									Scheduled	Successfully assigned openshift-l573a to 172.17.4.99
  2m		2m		1	{kubelet 172.17.4.99}	implicitly required container POD			Pulled		Container image "gcr.io/google_containers/pause:0.8.0" already present on machine
  2m		2m		1	{kubelet 172.17.4.99}	implicitly required container POD			Created		Created with docker id c24533213fa7
  2m		2m		1	{kubelet 172.17.4.99}	implicitly required container POD			Started		Started with docker id c24533213fa7
  2m		2m		1	{kubelet 172.17.4.99}	spec.containers{origin}					Pulling		Pulling image "openshift/origin"
  33s		33s		1	{kubelet 172.17.4.99}	spec.containers{origin}					Pulled		Successfully pulled image "openshift/origin"
  33s		33s		1	{kubelet 172.17.4.99}	spec.containers{origin}					Created		Created with docker id 71d378018e83
  33s		33s		1	{kubelet 172.17.4.99}	spec.containers{origin}					Started		Started with docker id 71d378018e83


