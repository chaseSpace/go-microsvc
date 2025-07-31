# Go微服务模板

一个简洁、清爽，且不使用任何**微服务框架**的微服务项目架构，从变量命名到不同职责的（多层）目录结构定义。

**目录**

<!-- TOC -->
* [Go微服务模板](#go微服务模板)
  * [1. 启动&停止日志输出](#1-启动停止日志输出)
  * [2. 目录结构释义](#2-目录结构释义)
  * [3. 如何使用](#3-如何使用)
  * [4. 本地（dev）环境启动微服务的原理](#4-本地dev环境启动微服务的原理)
  * [5. 工具下载（更新）](#5-工具下载更新)
    * [5.1 下载protoc](#51-下载protoc)
    * [5.2 下载protoc插件](#52-下载protoc插件)
  * [6. 其他](#6-其他)
    * [说明](#说明)
    * [建议](#建议)
    * [拉取代码](#拉取代码)
    * [部署&开发须知](#部署开发须知)
    * [资源链接](#资源链接)
<!-- TOC -->

> **完成进度：100%**

支持以下模式或特性：

- ✅ 使用单仓库多服务模式
- ✅ 使用grpc+protobuf作为内部rpc通讯协议
- ✅ 统一API Gateway管理南北流量
    - ✅ 透明转发HTTP流量到后端服务，无编码开销
    - ✅ 能够动态转发流量至新增服务，无需重启
- ✅ 支持 Kubernetes/PM2 部署方式
- ✅ 使用K8s Service作为服务发现
- ✅ 使用gorm作为orm组件，支持扩展
- ✅ 使用Redis作为cache组件，支持扩展
- ✅ 支持使用Redis/Kafka作为消息队列
- ✅ RPC超时重试、熔断和全局限速（server侧）功能
- ✅ 支持本地开发环境启动**多个**微服务，且支持RPC调用（不依赖注册中心组件）

其他有用的特性：

- ✅ 接入外部IM告警（钉钉）
- ✅ 接入三方平台功能（SMS、审核等，详见[第三方SDK接入.md](docs/%E7%AC%AC%E4%B8%89%E6%96%B9SDK%E6%8E%A5%E5%85%A5.md)）
- ✅ shell脚本支持多平台环境
- ✅ 定义微服务Error类型，以便跨服务传递error（在gRPC拦截器中解析），[查看代码](./pkg/xerr/err.go)
- ✅ 跨多个服务传递metadata示例（通过Context），搜索函数`TraceGRPC`
- ✅ gRPC Client 拦截器示例，包含`GRPCCallLog`, `ExtractGRPCErr`, `CircuitBreaker`, `Retry`, `WithFailedClient`
- ✅ gRPC Server
  拦截器示例，包含`RecoverGRPCRequest`, `ToCommonResponse`, `LogGRPCRequest`, `TraceGRPC`, `StandardizationGRPCErr`
- ✅ 各微服务在Interceptor实现JWT+Cache鉴权

运行通过的示例：

- ✅ **本地**单服务GRPC接口测试用例（[user-ext_api_test](./test/user/ext_test.go)）
- ✅ **本地**跨服务GRPC调用测试用例（[admin-ext_api_test](./test/admin/ext_test.go)）
- ✅ **Gateway** HTTP接口测试（调用后端微服务），使用Goland `HTTP Request`
  功能（[apitest_user.http](./test/gateway/apitest_user.http)）

目前已提供常见的微服务示例：

- admin: 管理后台（非完整）
- user：用户模块（实现基础的注册、登录功能）
- currency：资产模块（包含流水、消费、进账的货币功能）
- thirdparty：第三方模块（接入各类第三方服务，如审核、短信和认证等）

本项目文档指引：

- [使用证书加密以及指定授权gRPC通信](docs/开发必看/generate_cert_for_svc.md)

## 1. 启动&停止日志输出

略。

## 2. 目录结构释义

TODO

## 3. 如何使用

```shell
git clone --recurse-submodules https://github.com/chaseSpace/go-microsvc.git
cd go-microsvc/
go mod download

# 启动服务
go run service/user/main.go
go run service/admin/main.go
go run service/gateway/main.go
...

# 调用gateway，参考 根/test/gateway/apitest.http
```

本项目已支持在K8s环境中部署，请参考[在K8s上部署此项目](docs/开发必看/deploy_on_k8s.md)。

## 4. 本地（dev）环境启动微服务的原理

理论上来说，调用微服务是走注册中心的，要想在本地启动多个微服务且能正常互相调用，又不想在本地部署一个类似etcd/consul/zookeeper
的注册中心，最简单的办法是：

实现一个简单的注册中心模块，然后**在开发环境**随服务启动。

- [~~网络协议之mDNS~~（由于Windows支持不完善，不再使用）](https://www.cnblogs.com/Alanf/p/8653223.html)
- [simple_sd实现](./xvendor/simple_sd)

注意：dev环境启动的微服务仍然连接的是**beta环境的数据库**。

## 5. 工具下载（更新）

### 5.1 下载protoc

工具以及插件的二进制文件都已经包含在本仓库的`tool/`目录下，使用项目的每个人在拉取项目后都无需另外下载。
已下载的是protoc v24.4版本，其余插件也是编写本项目时的最新版本（下载时间更新至2023年10月5日）。

如需更换版本（所有平台都要换），可点击下方链接自行下载：

https://github.com/protocolbuffers/protobuf/releases

### 5.2 下载protoc插件

本仓库的`tool/`,`tool_mac/`都已经包含这些插件，这里只是演示如何下载，你可以使用这个方式来更新插件版本。

```shell
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.16
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@v2.16

# 检查是否下载成功
[root@localhost go-microsvc-template]# ls $GOPATH/bin/
protoc         protoc-gen-go-grpc     protoc-gen-grpc-gateway  protoc-gen-swagger
protoc-gen-go  protoc-gen-openapiv2   

# 下载后需要复制到仓库下的tool目录（以及tool_mac），其他人拉取代码后，无需再下载
cp $GOPATH/bin/* tool/protoc_v24
```

若要更改版本，建议同时修改`tool/proto_v24/`目录名称，并同步修改`build_pb.sh`脚本中对该目录的引用部分，以便更新版本后脚本能够正常运行。

> 在Windows上运行shell脚本，你需要为系统配置`bash`解释器。

## 6. 其他

### 说明

1. 增加微服务时，需要重新颁发微服务的 [client证书](deploy/dev/cert/client-cert.pem)
   ，务必按照[文档](docs/开发必看/generate_cert_for_svc.md#232-生成client证书)进行操作。
2. 开发常用库
    - 泛型库：lo
    - 颜色打印: pp
    - 网络请求：gorequest
    - 唯一id：ksuid
    - 类型转换：cast
    - 支持分组的正则：regexp2
    - UUID：go-uuid
    - 配置读取：viper
    - 测试：testify
    - 进程缓存：patrickmn/go-cache

### 建议

- `protocol/`是存放生成协议代码的目录，在实际项目开发中可以加入`.gitignore`文件，以避免在PR review时产生困扰；
- 单独保存用于办法微服务证书的根证书私钥[ca-key.pem](deploy/dev/cert/ca-key.pem)，勿提交到代码仓库；
- 本项目中的admin（管理后台）服务没有实现具体功能，建议实际项目中使用开源的前后端配套admin项目；

### 拉取代码

项目中的`proto/`目录是git submodule，所以在拉取时需要注意：

```shell
# 若是第一次拉取项目，则执行
git clone --recurse-submodules {git_addr}

# 若是拉取了项目，但没有拉取submodule，则执行
git submodule update --init

# 若要提交子模块，进入子模块切换主分支（仅首次本地commit时需要），然后commit即可；
# -- 主项目与子模块是以commit-id形式关联的
```

拉取成功后，`proto`目录下才会有内容，然后可以执行`make pb`命令来生成protobuf代码（目标文件在`protocol/`目录下），此时各微服务可以编译。

### 部署&开发须知

- [项目文档](docs)
- 开发规范：
    - PB枚举定义时，一般情况下勿使用0值
    - 新增数据表时遵循【先写DDL-SQL，再写model代码】的顺序规范；
        - 所有的建表SQL均在 [docs/model_sql](docs/model_sql) 目录下，以`*.sql`文件形式存放；
        - 所有表model代码均在 [proto/model](proto/model) 目录下；
    - 单个服务内的目录结构已经包括各种场景，均可参考`user`服务，勿自定义以便协作维护
    - 后端分页查询搜索 `db.PageQuery`
- 实际项目中，考虑去除项目中`TEMP TEST`关键字处的临时代码

### 资源链接

- [K8s中文教程](https://github.com/chaseSpace/k8s-tutorial-cn/tree/main)
- [Consul 官网介绍](https://developer.hashicorp.com/consul/docs/intro)
- [Consul 服务发现原理](https://developer.hashicorp.com/consul/docs/concepts/service-discovery)
