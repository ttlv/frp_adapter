![golang](https://img.shields.io/badge/golang-1.14.0-green.svg?style=plastic)

# 1. 应用背景

在传统的Cloud与Edge的架构中，Edge Side基于NAT(Network Address Translation)网络，边缘节点集群可绑定一个公网IP，可以实现边到云的访问。但是在实际场景中，很多时候都需要实现从云端直接访问边缘端的需求，比如说运维人员需要在某些时候ssh登陆到指定的边缘节点进行操作。现有比较成熟的就是FRP内网穿透方案。从而解决从外网访问内网机器的需求。

# 2. FRP简介

Frp 是一个专注于内网穿透的高性能的反向代理应用，支持 TCP、UDP、HTTP、HTTPS 等多种协议。可以将内网服务以安全、便捷的方式通过具有公网 IP 节点的中转暴露到公网。并且Frp是一个C/S架构的服务，需要有服务端与客户端，服务端即Frp Server(以frps简称),客户端即Frp Client(以frpc简称)。通过在具有公网 IP 的节点上部署 frp 服务端，可以轻松地绕过防火墙的限制穿透内网，同时提供诸多专业的功能特性，这包括：

* 客户端服务端通信支持 TCP、KCP 以及 Websocket 等多种协议。
* 采用 TCP 连接流式复用，在单个连接间承载更多请求，节省连接建立时间。
* 代理组间的负载均衡。
* 端口复用，多个服务通过同一个服务端端口暴露。
* 多个原生支持的客户端插件（静态文件查看，HTTP、SOCK5 代理等），便于独立使用 frp 客户端完成某些工作。
* 高度扩展性的服务端插件系统，方便结合自身需求进行功能扩展。
* 服务端和客户端 UI 页面。

# 3. Frp Adapter(使用Frpa为简称)

## 3.1 Frpa简介

Frp Adapter作为介于k8s与Frps的中间HTTP服务，诞生的目的就是为了解决Frpc与Frps的状态维护的问题，基于现有的业务场景，Frps与Frpc的状态对于业务来说非常重要，所以需要有一个中间的服务可以承上启下去对接Frps与k8s集群，实时的去维护Frpc与Frps的状态。维护状态并未使用关系型数据库，而是将状态数据直接存储到ETCD中，从而设计了一套CRD(ps:frpa的意义就是维护CRD的状态)虽然这有悖于k8s声明式API的设计思想,对于业务而言，这还是一种比较合适的实现方式。注意虽然定义了CRD但是并没有配套的controller，Frp Adapter本身就是一种类似controller的方式，只不过是以微服务的方式存在。

## 3.2 Frpa功能

### 3.2.1 维护Frpc与Frps的状态

#### 1. /frp_fetch/:nm_name

#####  请求方式

GET

##### 参数传递

nm_name: NM对象的名字(具体格式为nodemaintenances-xxxx，xxxx是unique_id)

##### api功能介绍

获取指定名字的NM对象

#### 2. /frp_create

##### 请求方式

POST

##### 请求参数

frp_server_ip_address: frps的公网IP地址

unique_id: frpc mac地址经过hash计算的出来的唯一值

port: frpc与frps的交互端口号

##### api功能介绍

负责创建NM对象

#### 3. /frp_update

##### 请求方式

PUT

##### 请求参数

frp_server_ip_address: frps的公网IP地址

status: frpc的状态,只有两个枚举值,online和offline

unique_id: frpc mac地址经过hash计算的出来的唯一值

port: frpc与frps的交互端口号

##### api功能介绍

更新NM对象

#### 4. /nm_useless

##### 请求方式

PUT

##### 请求参数

无

##### api功能介绍

使k8s集群中所有NM对象无效。

### 3.2.2 ssh反向代理

#### 1. /reverse_proxy/:nm_name

##### 请求方式

WebSocket

##### 请求参数

nm_name: NM对象的名字(具体格式为nodemaintenances-xxxx，xxxx是unique_id)

##### api功能介绍

web或者是cli以websocket的形式发起请求，建立连接，以ssh方式远程登录到目标节点

#### 2. /reverse_proxy_shell

##### 请求方式

POST

##### 请求参数

nm_name: NM对象的名字(具体格式为nodemaintenances-xxxx，xxxx是unique_id)

cmd: 待执行的shell命令的字符串

##### api功能介绍

反向代理建立ssh连接执行shell命令

# Frp Adapter架构设计图

![Frp Adapter 架构图](https://images-1253546493.cos.ap-shanghai.myqcloud.com/frpa.jpg)



# 4. FRPA安装方式

### 1. pod形式安装

#### 1. 创建deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: frp-adapter
spec:
  replicas: 1
  selector:
    matchLabels:
      app: frp-adapter
  template:
    metadata:
      labels:
        app: frp-adapter
    spec:
      containers:
      - name: frp-adapter
        image: gopherlv/frp-adapter:frp-adapter-48821173
        env:
          - name: FRP_ADAPTER_ADDRESS
            value: :8888
          - name: FRP_SERVER_HTTPAUTHUSERNAME
            value: admin
          - name: FRP_SERVER_HTTPAUTHPASSWORD
            value: admin
          - name: FRP_SERVER_API
            value: http://xxxxx:7500/api/proxy/tcp // 替换成frps服务的ip地址
        imagePullPolicy: Always
        securityContext:
          privileged: true
      nodeSelector:
        kubernetes.io/hostname: xxxxxx 自行打label
      restartPolicy: Always

```

#### 2. 创建svc

```yaml
apiVersion: v1
kind: Service
metadata:
  name: frpa-service-nodeport
spec:
  selector:
      app: frp-adapter
  ports:
    - name: http
      port: 8888
      protocol: TCP
      targetPort: 8888
  type: NodePort

```

### 2. bin直接运行

```shell
source dev_env
cd main && go build -o frp-adapter
./frp-adapter
```

## 5. 环境变量参数

```shell
export FRP_ADAPTER_ADDRESS=":8888" // frp-adapter http server port
export FRP_SERVER_HTTPAUTHUSERNAME="admin" // frps http auth username
export FRP_SERVER_HTTPAUTHPASSWORD="admin" // frps http auth password
export FRP_SERVER_API="http://10.1.11.38:7500/api/proxy/tcp" // frps server api
```

## 6. 注意事项

Frps与Frpc的均基于0.33版本的Frp源码进行了一定程度的修改，想要使用Frp Adapter请参考修改以后的Frps宇Frpc的源码

[适配Frp Adapter的Frp的源码版本,注意是master分支](https://github.com/ttlv/frp )

[fatedier Frp的源码传送门](https://github.com/fatedier/frp)

[NodeMaintenances CRD定义](https://github.com/ttlv/nodemaintenances)

## 7. 开发者

* [ttlv](http://github.com/ttlv)
* [kkBill](https://github.com/kkBill)


## 贡献代码

非常欢迎优秀的开发者来贡献Frp Adapter。在提Pull Request之前，请首先阅读源码，了解原理和架构。如果不懂的可以加他的微信 `handsomett950602` 注明 `Frp Adapter`。

## 社区

**在此特别鸣谢 Frp的作者,浙大SEL实验室以及杭州谐云科技有限公司提供的支持。**

如果您觉得 Frp Adapter对您有帮助，请扫描下方二维码，如果无法添加请加微信 `handsomett950602` 并注明`Frp Adapter 开源交流`,欢迎各位老板赏咖啡支持微信支付和比特币充值，欢迎提issue。

<p align="center">
    <img src="https://ocx.oss-cn-hangzhou.aliyuncs.com/dev/coinx-admin/banners/1/we_chat_scan_code.jpg" height="360">
    <img src="https://ocx.oss-cn-hangzhou.aliyuncs.com/dev/coinx-admin/banners/1/we_chat_pay_scan_code.jpeg" height="360">
    <img src="https://ocx.oss-cn-hangzhou.aliyuncs.com/dev/coinx-admin/banners/1/btc_deposit_scan_code.jpg" height="360">
</p>

