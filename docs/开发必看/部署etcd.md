## 部署etcd

若使用 etcd 作为服务注册中心，则参考此文档部署。

### 1. 开发环境

```shell
# 单节点最简命令（v3.5.21）,一般使用次新主版本（最新稳定版本是 3.6）
# Note: 2379是client连接端口，2380是组集群时节点之间的连接端口
docker run -d --name etcd \
  -p 2379:2379 \
  gcr.io/etcd-development/etcd:v3.5.21 \
  /usr/local/bin/etcd \
  --name node1 \
  --listen-client-urls http://0.0.0.0:2379 \
  --advertise-client-urls http://127.0.0.1:2379
```

**验证**

```shell
# 看集群成员
docker exec -it etcd etcdctl --endpoints=http://localhost:2379 member list

# 看集群健康
docker exec -it etcd etcdctl --endpoints=127.0.0.1:2379 endpoint health
```