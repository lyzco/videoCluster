# VidCluster

[![license](https://img.shields.io/github/license/apache/incubator-seata-go.svg)](https://www.apache.org/licenses/LICENSE-2.0.html)

**分布式视频聚合与流媒体系统**

[English README 🌐](./README.md)

---

## 📦 项目概述

VidCluster 是一个支持多节点视频存储、聚合与流媒体播放的分布式系统。

- 🚀 基于 Go 和 Kratos 框架开发
- 🛰️ 具备良好的可扩展性与可维护性
- 🔐 支持鉴权与权限管理（后续功能）

---

## 🎯 主要特性

- 多节点分布式视频服务
- 使用 gRPC 进行节点管理
- 支持 HTTP Range 请求进行视频流播放
- 中央服务统一聚合与索引视频元数据
- 支持未来横向扩展

---

## 🏗 系统架构

- **中央服务**
    - 视频聚合
    - 用户鉴权
    - 视频权限控制
    - 节点注册与管理
    - 播放列表管理

- **节点服务**
    - 本地视频元数据服务
    - HTTP 视频流服务

---

## ⚙ 技术栈

- Golang
- Kratos
- gRPC
- HTTP (Range Streaming)
- Consul

---

## 📂 项目结构


---

## 🔒 许可协议

VideoCluster 使用 Apache 许可证2.0版本，请参阅 LICENSE 文件了解更多。

