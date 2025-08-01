## 在K8s上部署此项目

因为K8s原生支持使用DNS作为服务（Pod）的服务发现方式，所以如果使用K8s部署此项目，我们无需再部署类似Consul/Etcd等组件用于服务注册/发现。

本文档介绍如何使用Helm在K8s上部署此项目。

首先，本项目已经使用`helm create CHART_NAME`创建了一个Helm chart，位于`根/deploy/service-chart`。这是一个通用的微服务chart，
在部署时会引用每个微服务目录中的`values.yaml`中的配置（例如`service/user/helmValues.yaml`），并使用`templates`
目录下的模板文件生成K8s资源。

也就是说，所有微服务共用于一套helm chart模板，各自管理自己的`values.yaml`即可。

> 你可以查看 [Helm手记](https://github.com/chaseSpace/k8s-tutorial-cn/blob/main/doc_helm.md)
> 来快速入门Helm，[下载地址](https://github.com/helm/helm/releases)。


部署步骤如下：

- 构建服务镜像并推送至仓库
- 部署服务的helm chart

下面的步骤以部署dev环境的user和gateway服务为例：

```shell
SVC=user
ENV=beta
TAG=1.0.0

# 在shell中进入项目根目录，执行构建脚本（构建测试环境包）
sh build_image.sh $SVC $ENV $TAG

# 验证helm模板无误
helm lint deploy/go-svc-chart -f ./service/$SVC/helmValues.yaml

# 检查helm模板
helm install --dry-run --debug example deploy/go-svc-chart -f ./service/$SVC/helmValues.yaml \
  --set image.tag=$TAG --set variables.deployEnv=$ENV --description="image: $TAG"

# 测试部署(若只有一台机，则需要删除污点，参考下面的方式)
helm install go-$SVC ./deploy/go-svc-chart -f ./service/$SVC/helmValues.yaml \
  --set image.tag=$TAG --set variables.deployEnv=$ENV --description="image: $TAG"

# 查看helm部署（APP VERSION可以暂时忽略）
$ helm ls                                                                         
NAME        NAMESPACE	REVISION	UPDATED                                	STATUS  	CHART             	APP VERSION
go-user   	default  	1       	2023-12-05 22:56:08.625016541 +0800 CST	deployed	go-svc-chart-0.1.0	1.0.0

# 查看chart创建的k8s资源及其状态
$ helm status go-$SVC --show-resources

# 以同样的步骤部署gateway
```

开发/测试环境可能只有一个k8s节点，需要删除master节点污点才能正常运行pod：

```shell
$ kubectl describe node k8s-master | grep Taints
Taints:             node-role.kubernetes.io/control-plane:NoSchedule
$ kubectl taint node k8s-master node-role.kubernetes.io/control-plane:NoSchedule-
node/k8s-master untainted
```

访问服务：

```shell
$ kubectl get svc                                                    
NAME         TYPE        CLUSTER-IP     EXTERNAL-IP   PORT(S)    AGE
go-gateway   ClusterIP   20.1.235.201   <none>        8000/TCP   103m
go-user      ClusterIP   20.1.171.170   <none>        3000/TCP   97m
kubernetes   ClusterIP   20.1.0.1       <none>        443/TCP    16d

# 通过service ip访问gateway
$ curl 20.1.235.201:8000/ping                                 
pong
# 通过service ip访问user的注册接口（gateway转发）
$ curl -X POST  http://20.1.235.201:8000/forward/svc.user.UserExt/SignUpAll -d {}
{"code":400,"msg":"无效昵称或超出长度","data":null}

# 若要通过ingress访问
# 先获取ingress控制器的svc ip和端口，这里是20.1.166.146:30189
$ kubectl get svc -ningress-nginx
NAME                                 TYPE           CLUSTER-IP     EXTERNAL-IP   PORT(S)                      AGE
ingress-nginx-controller             LoadBalancer   20.1.166.146   <pending>     80:30189/TCP,443:30415/TCP   10d
ingress-nginx-controller-admission   ClusterIP      20.1.75.220    <none>        443/TCP                      10d

# 然后访问节点的30189端口会自动转发至nginx控制器的80端口
$ curl 127.0.0.1:30189/ping                                 
pong
$ curl 127.0.0.1:30189/forward/svc.user.UserExt/SignUp -d {}
{"code":400,"msg":"无效昵称或超出长度","data":null}
```

### 配置gateway的域名和证书

ingress在前面已经随gateway自动安装好了，检查一下：

```shell
$ kk get ingress                 
NAME         CLASS   HOSTS              ADDRESS   PORTS     AGE
go-gateway   nginx   go-gateway.micro             80, 443   15m

$ kk describe ingress go-gateway
```

ingress内引用了名为`go-gateway-tls`的Secret，用来保存证书和域名，
但证书还没有，可参考[此文档](generate_cert_for_svc.md31-生成server证书) 创建网关证书（仅用于测试）；
然后再将证书导入Secret：

```shell
cd deploy/beta/cert/
kk create secret tls go-gateway-tls --cert=gateway-cert.pem --key=gateway-key.pem
```

### 在外部使用HTTPS域名访问

预期的访问链路：宿主机（模拟外部环境）--ingress控制器--Service(gateway)--子服务

ingress控制器应该已经随集群安装好，看一下：

```shell
$ kk get svc -ningress-nginx
NAME                                         TYPE           CLUSTER-IP     EXTERNAL-IP   PORT(S)                      AGE
service/ingress-nginx-controller             LoadBalancer   20.1.175.255   <pending>     80:30507/TCP,443:31435/TCP   4h18m
service/ingress-nginx-controller-admission   ClusterIP      20.1.242.90    <none>        443/TCP                      4h18m
```

ingress-nginx默认的Service安装方式是LoadBalancer，它的端口是映射到宿主机的端口，如上面示例中的 `80:30507` (**Service:
NodePort**)，以及`443:31435`。

这里我们使用k8s节点的宿主机来模拟“外部”，先配置hosts文件：

```shell
tlsPort=$(kubectl get svc -ningress-nginx ingress-nginx-controller -o jsonpath='{.spec.ports[1].nodePort}')
echo "127.0.0.1 go-gateway.micro" >> /etc/hosts
```

访问`gateway`：

```shell
$ pwd
/mnt/hgfs/gg/go-microsvc-template/deploy/dev/cert

# 任何请求都经过 ingress-nginx 控制器处理
$ curl --cacert ./ca-cert.pem https://go-gateway.micro:31435     
<html>
<head><title>404 Not Found</title></head>
<body>
<center><h1>404 Not Found</h1></center>
<hr><center>nginx</center>
</body>
</html>

# 网关处理ping请求
$ curl --cacert ./ca-cert.pem https://go-gateway.micro:$tlsPort/ping
pong
# 调用子服务接口（gateway转发），这里的错误提示是子服务检查参数不通过，说明请求已经透传到子服务
$ curl --cacert ./ca-cert.pem https://go-gateway.micro:31435/forward/svc.user.UserExt/SignUpAll -d {}
{"code":400, "msg":"ErrParams ➜ missing field:`base`", "msg_chinese":"", "from_gateway":false, "data":null}
```

### 进入Pod调试

如果需要的话，可进入Pod进行调试：

```shell
# 子服务使用alpine 镜像，缺少常用指令，--image指定其他镜像来调试，可选择 busybox, curl等
kk debug go-gateway-xxx -it --image appropriate/curl -- sh
```

### 发布升级

此后，若要更新服务，按下面的步骤进行：

- 修改代码（推送到仓库）
- 构建镜像->推送镜像到仓库（生产环境要修改tag）
- 执行Helm升级命令

部分命令示例：

```shell
# 以user服务为例
SVC=user
TAG=1.0.1

# 开发/测试环境迭代
helm upgrade go-$SVC ./deploy/go-svc-chart -f ./service/$SVC/helmValues.yaml --description="image: $TAG"
# 查看历史修订版本REVISION
helm history go-$SVC
# 回滚到指定 REVISION
helm rollback go-$SVC 1

# 生产环境迭代（指定tag）
helm upgrade go-$SVC ./deploy/go-svc-chart -f ./service/$SVC/helmValues.yaml --set image.tag=$TAG --description="image: $TAG"
```

> Helm的APP Version暂时没有好的办法在部署时自动修改，所以使用`--description`选项记录每次升级使用的镜像版本

### 其他命令

发布后可能需要检查当前发布使用的values配置，可以通过下面的命令查看：

```shell
helm get values go-user
helm get values go-user --revision 1 #查看指定REVISION的values
helm un go-user # 删除服务
```

### 故障解决

如果你使用Docker HUB作为镜像仓库，可能会因为网络问题导致Pod一直在拉取中或者拉取失败，可以尝试手动拉取：

```shell
ctr -n k8s.io img pull docker.io/leigg/go-user:1.0.0
```

实际项目中应该使用国内托管仓库或者自建仓库，以确保稳定拉取。