## 部署consul

若使用Consul作为服务注册中心，则参考此文档部署 consul。

### 1. 开发环境

| **配置项**   | **推荐设置**                                                                                                         |
|-----------|------------------------------------------------------------------------------------------------------------------|
| **运行模式**  | `-dev`（开发者模式，自动配置单节点集群）                                                                                          |
| **启动命令**  | `consul agent -dev` 或 `docker run -d --name=consul -p 8500:8500 hashicorp/consul agent -dev -ui -client=0.0.0.0` |
| **节点类型**  | 自动为 server 模式（单节点，无 client）                                                                                      |
| **数据持久化** | 可选（默认使用临时目录，重启后数据丢失）                                                                                             |
| **端口暴露**  | 仅 8500（HTTP API 和 UI）即可                                                                                          |
| **网络要求**  | 本地回环即可，无跨节点通信需求                                                                                                  |
| **配置复杂度** | 极低，无需配置文件                                                                                                        |
| **适用场景**  | 本地开发、功能测试、学习 Consul API                                                                                          |

> ⚠️ 警告：禁止在生产环境使用 -dev 模式，因为它不安全、不支持多节点、无持久化。

若在Linux环境运行，请使用：

```shell
# 其中 --add-host 部分是为了允许 consul 访问宿主机服务
docker run -d \
  --name consul \
  -p 8500:8500 \
  --add-host=host.docker.internal:host-gateway \
  hashicorp/consul agent -dev -ui -client=0.0.0.0

```

### 2. 生产环境

| **配置项**    | **推荐设置**                                                                                                         |
|------------|------------------------------------------------------------------------------------------------------------------|
| **运行模式**   | **Server 模式**（3-5 个节点组成法定仲裁）                                                                                     |
| **客户端节点**  | 每个运行服务的主机部署一个 **Client 模式** 代理                                                                                   |
| **启动命令示例** | `consul agent -server -bootstrap-expect=3 -data-dir=/opt/consul -ui`                                             |
| **数据持久化**  | 必须挂载持久化目录（如 `/opt/consul`）到宿主机或 Docker volume                                                                    |
| **端口配置**   | 需开放以下端口：<br>- 8300（Server RPC）<br>- 8301（LAN Gossip）<br>- 8302（WAN Gossip）<br>- 8500（HTTP API/UI）<br>- 8600（DNS） |
| **网络要求**   | 所有节点间 RTT < 50ms（平均），99% 流量 RTT < 100ms                                                                          |
| **节点发现**   | 使用 `-retry-join` 或 `auto-join` 自动加入集群（如 EC2 tag、Kubernetes 等）                                                    |
| **配置文件**   | 使用 `HCL` 或 `JSON` 配置文件（如 `/etc/consul.d/server.hcl`）                                                             |
| **高可用配置**  | 至少 3 个 server 节点，避免偶数节点（防止脑裂）                                                                                    |
| **安全配置**   | 启用 ACL、TLS 加密、Gossip 加密（非开发环境必须）                                                                                 |
| **监控与日志**  | 配置日志级别、Prometheus 指标、健康检查                                                                                        |
| **升级与维护**  | 支持滚动升级，使用 `consul leave` 优雅下线节点                                                                                  |

参考：https://developer.hashicorp.com/consul/docs/fundamentals/agent

### 3. client与server节点的区别

| 对比维度      | Server 节点                | Client 节点               |
|-----------|--------------------------|-------------------------|
| **角色定位**  | 控制平面：存储集群状态、Raft 共识、选主   | 数据平面：代理转发、轻量级无状态        |
| **是否持久化** | ✅ 本地持久化 Catalog 与 KV 数据  | ❌ 仅缓存，不持久化              |
| **资源占用**  | 高（CPU/内存/磁盘）             | 低（仅网络与少量 CPU）           |
| **数量建议**  | 3-5 个（奇数，避免脑裂）           | 可随业务节点无限扩展              |
| **功能范围**  | 服务注册、发现、健康检查、选主、复制       | 健康检查、转发请求、本地缓存          |
| **网络端口**  | 8300/8301/8302/8500/8600 | 8301/8500/8600（无需 8300） |
| **故障影响**  | 影响集群选主及一致性               | 仅影响本机服务发现               |
| **部署位置**  | 独立、稳定、高可用主机              | 与业务应用同机部署（每台机器一个）       |
