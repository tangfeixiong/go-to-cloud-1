
[vagrant@localhost ~]$ sudo password vagrant
sudo: password：找不到命令
[vagrant@localhost ~]$ sudo passwd vagrant

[vagrant@localhost ~]$ vim .bashrc 
[vagrant@localhost ~]$ export PATH=/data/bin:$PATH


[vagrant@localhost origin]$ cp _output/local/bin/linux/amd64/* /data/bin


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



[vagrant@localhost ~]$ vim openshift-controller.yaml 

[vagrant@localhost ~]$ kubectl --kubeconfig=coreos.kubeconfig create -f openshift-controller.yaml 
replicationcontroller "openshift" created

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


## QingYuan office

[centos@k8s-master-ha-2 amd64]$ ./oc --config=/home/centos/.kube/config secrets new openshift-local-config /home/centos/devops/amd64/openshift.local.config/master -o json | tee secret.json
error: the server could not find the requested resource
See 'oc secrets new -h' for help and examples.


