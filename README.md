# whimer

![golangci-lint](https://github.com/ryanreadbooks/whimer/workflows/golangci-lint/badge.svg)
![govulncheck](https://github.com/ryanreadbooks/whimer/workflows/govulncheck/badge.svg)

Whimer 是一个面向内容社区场景的微服务后端系统，覆盖用户、内容、关系、评论、搜索、消息和实时推送能力。

## 功能模块

| 服务 | 对外接口 | 核心功能 |
|------|----------|----------|
| **pilot** | HTTP | API 聚合网关，编排下游 gRPC 服务 |
| **passport** | HTTP + gRPC | 用户信息，登录鉴权等 |
| **note** | gRPC | 笔记内容：笔记发布、点赞等|
| **comment** | gRPC | 笔记评论：评论发表、点赞、回复等 |
| **relation** | gRPC | 用户关系：关注、粉丝、关注列表、粉丝列表等 |
| **search** | gRPC | 关键词/标签搜索 |
| **counter** | gRPC | 通用计数能力：承载点赞等能力 |
| **msger** | gRPC | IM消息，包括系统消息，用户单聊等 |
| **wslink** | WebSocket + gRPC | 长连接管理 |
| **conductor** | gRPC | 任务编排与调度 |
| **lambda/media** | - | 图片/视频处理：图片压缩，视频转码等|
| **s1-gate** | HTTP | 对象存储服务网关，处理资源访问等，基于OpenResty|

## 服务入口

**直接面向客户端的服务：**
- **pilot** - 主 HTTP API 网关，大部分业务接口聚合入口
- **passport** - 用户、登录与认证相关 HTTP 接口
- **wslink** - WebSocket 长连接，用于实时推送
- **s1-gate** - 对象存储服务网关，基于 OpenResty，代理资源访问

**内部服务（仅 gRPC）：**
- note、comment、relation、search、counter、msger、conductor

## 通信方式

- **同步**：服务间通过 gRPC 调用
- **异步**：Kafka 用于笔记事件、搜索索引更新、系统通知等场景
- **任务编排**：conductor 负责任务调度，lambda/media 作为 Worker 处理媒体任务

## 存储与基础设施

- **MySQL** 
- **Redis** 
- **Elasticsearch** 
- **Kafka** 
- **MinIO** 
- **Etcd** 
- [**imgproxy**](https://github.com/imgproxy/imgproxy) 
- [**go-zero**](https://github.com/zeromicro/go-zero) 
- [**OpenResty**](https://github.com/openresty/openresty) 
