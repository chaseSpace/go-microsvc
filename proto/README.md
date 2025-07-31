# rpc-protocol

此原本为后端微服务使用的基于 protobuf 的 RPC 协议，但后端架构已通过网关将HTTP请求转换为RPC请求，所以前端开发人员也可参照此项目来编写
HTTP 请求代码。

### 0. 拉取

这个模块是前后端共用，所以是作为 submodule 引入到项目里，**初次引入**通过以下方式：

```shell
# 假设现在已经在你的Git项目内（web前端 或 移动端）
git submodule add git@github.com:CocktailPub/rpc-protocol.git proto
git add proto
git commit -m "add rpc-protocol"
```

此时，你所在路径下应该有`proto`文件夹存在。拉取后只需更新即可：

```shell
git submodule update --remote
```

- [关于 protobuf](https://protobuf.com.cn/overview/)
- [VSCode or IDEA 安装protobuf插件](https://juejin.cn/post/7357569755966996490)

**注意**：前端开发人员一般不需要修改此仓库代码，待后端开发人员更新后拉取即可。

### 1. 目录说明

前端需要关心的目录（文件）是`svc/*/*.proto`，下面通过2个子目录来说明：

```shell
├── barbasepb # barbase（酒吧）子服务的协议
│   ├── const.proto
│   ├── ext.proto # 外部接口协议（前端关心）
│   └── int.proto # 内部接口协议（前端忽略）
├── commonpb # 公共协议，可以被其他子服务引用
│   ├── admin.proto
│   ├── common.proto
│   ├── enum.proto
│   ├── thirdparty.proto
│   └── user.proto
```

### 2. 协议文件说明

调用后端接口时，由于后端是微服务架构，所以前端首先需要了解到以下信息，接口所属服务和接口名称，例如：

- 接口所属服务：barbase
- 接口名称：ListRecommendBars

> 对于管理后台，后端限制只能调用`admin`服务外部（Ext）接口；客户端则只能调用`admin`以外服务的外部（Ext）接口。

然后，我们可以在`svc/barbasepb/ext.proto`中查看到接口定义，每个`ext.proto`文件的结构如下：

```
syntax = "proto3";

message ListRecommendBarsReq {
    ...
}
message ListRecommendBarsRes {
    ...
}

...其他结构体定义

service barbaseExt {
  rpc ListRecommendBars(ListRecommendBarsReq) returns (ListRecommendBarsRes);
}
```

其中，`service`部分定义了这个服务下的所有接口，你可以从中找到接口名称，例如：ListRecommendBars。
显而易见，`ListRecommendBars`接口的请求参数是`ListRecommendBarsReq`，返回值是`ListRecommendBarsRes`,
你可以在该文件内找到这两个结构体的定义，例如：

```
message ListRecommendBarsReq{
  common.BaseExtReq base = 1;  # {类型} {字段名} = {编号};  除了int/string 这样的标准类型，其他类型都需要手动定义
  common.PageArgs page = 2;
}
message ListRecommendBarsRes{
  repeated Bar list = 1; # repeated 表示数组结构，元素类型是Bar
  int64 total = 2; // 用于计算页数
}
```

上述内容中的`common.BaseExtReq`是来自`svc/commonpb/common.proto`文件内定义的一个类型，你可以在文件开头的`import`
部分找到这种引用类型的定义位置，`common.PageArgs`同。

### 3. 调用示例

后端约定，所有HTTP请求的方法都是POST，请求头中包含`Content-Type: application/json`，
请求体和返回数据中包含JSON格式的请求参数。下面同样以`ListRecommendBars`为例进行说明。

#### HTTP 请求示例

假设我们的服务部署域名是`https://api.cocktail-hack.com`，那么 HTTP 请求可能如下：

请求方法: `POST`

请求路径: `/forward/barbase/ListRecommendBars`
，拼接格式为：`/forward/{服务名}/{接口名}`。

请求完整地址：`https://api.cocktail-hack.com/forward/barbase/ListRecommendBars`

请求头:

```
Content-Type: application/json
Accept: application/json
Authorization: {token} -- 鉴权接口必须（token通过登录/注册获取）
```

请求体:

```json
{
  "base": {
    "app_name": "cocktail-hack.pc",
    "app_version": "1.0.0",
    "platform": 3,
    "system": 4,
    "language": 2,
    "extension": {}
  },
  "page": {
    "pn": 1,
    "ps": 2,
    "is_download": false
  }
}
```

请注意，其中的`base`字段是所有对外接口的必需字段，前端可以将其进行封装，以便多接口调用。此外，
`base`中的`platform`, `system`，`language`都是枚举类型（int），需要关注对应的类型定义。

对于项目前期，暂且简化`base`传参，admin前端固定传参：

```
{
  "app_name": "admin-frontend",
  "app_version": "1.0.0",
  ...其他字段根据实际情况传入
}
```

#### HTTP 响应示例

状态码: 200 OK

响应头: `Content-Type: application/json`

响应体:

```json
{
  "code": 200,
  "msg": "OK",
  "data": {
    "list": [
      {
        "id": 1,
        "consume_note": "*For party of 5 or larger, a gratuity of 20% will be added to your bill automaticly.\nWe require each guest to order at least one drink. Thank you for understanding."
      },
      {
        "id": 6,
        "consume_note": "*prices stated are subjected to GST and service charge."
      }
    ],
    "total": 17
  }
}
```

注意：你可能会发现上面的**响应体**并不完全对应协议文件中的结构体`ListRecommendBarsRes`
的定义，其实这里也不难理解，因为`...Res`(Res代指Response)结构体是后端的单个RPC接口返回的数据格式定义，
而网关会对数据进行*标准化封装*，以方便前端读取和解析；简单来说，网关返回给前端的**统一**数据格式如下：

```json
{
  "code": 200,
  "msg": "OK",
  "data": {}
}
```

`data`部分承载的就是结构体`ListRecommendBarsRes`表示的数据，而`data.@type`是组件附加的对协议类型的描述（可忽略）。

当`code`为200时可以读取`data`，反之则应该打印或弹出错误提示`msg`。

## 4. 子模块的删除

如果你本地的子模块出现了混乱，并且很难再恢复正常，请执行删除再添加的步骤（不要push）。

```shell
git rm --cached proto
rm -rf proto

# 进入 .gitmodules 文件，删除 proto 相关的记录
vi .gitmodules

# 进入 .git/config 文件，删除 proto 相关的记录
vi .git/config

rm -rf .git/modules/proto
```

然后执行前面提到的命令，重新添加子模块。