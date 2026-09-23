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
- 扫描本地视频目录并使用 BadgerDB 保存元数据
- 使用“文件大小 + 前/中/后各 1 MiB”生成快速视频指纹
- 已存在的视频直接跳过缩略图生成和数据库写入
- Windows 使用系统 Shell 缩略图，Linux/macOS 使用 ffmpeg
- 支持未来横向扩展

> 当前 Node 服务的扫描、元数据查询和视频播放能力已经具备；Central
> 服务的节点发现与视频聚合仍在开发中。鉴权、权限控制和播放列表属于后续功能。

---

## 🔍 视频扫描流程

Node 服务扫描配置目录中的 `.mp4`、`.mkv`、`.avi`、`.mov` 和 `.flv`
文件，并按以下流程处理：

1. 读取文件大小，以及文件前部、中部、尾部各最多 1 MiB 的内容。
2. 使用 BLAKE3 为这些采样内容生成快速指纹，并将其作为视频 ID。
3. 如果 BadgerDB 中已经存在该 ID，则直接跳过。
4. 如果视频不存在，则按当前操作系统生成缩略图并保存元数据。

快速指纹最多读取约 3 MiB，适合大批量视频目录扫描，但它不是完整文件哈希：
只修改未采样区域且不改变文件大小时，指纹可能保持不变。

### 平台差异

| 平台 | 缩略图实现 | 运行要求 |
| --- | --- | --- |
| Windows | Windows Shell `IShellItemImageFactory` | 依赖系统已安装的视频编解码器，无需 ffmpeg |
| Linux / macOS | ffmpeg 截取视频帧 | `ffmpeg` 必须存在于 `PATH` 中 |

Windows 系统无法解析某种视频格式时，该视频仍会保存到数据库，但缩略图为空。

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
- ffmpeg
- BadgerDB

---

## ⚠️ 数据兼容性

早期版本使用完整文件 BLAKE3 Hash 作为视频 ID，当前版本改用采样快速指纹。
升级已有节点时，旧 ID 不会自动迁移；开发环境可以清理旧的 `*.data` 数据目录，
生产数据则需要先执行迁移。

---

## 📂 项目结构

```text
api/                  Central 和 Node 的 protobuf API
cmd/central/          Central 服务入口及 Wire 装配
cmd/node/             Node 服务入口及 Wire 装配
internal/central/     Central 的数据、服务与传输层
internal/node/        Node 的业务、数据、媒体处理、服务与传输层
internal/conf/        两个服务共用的配置定义
internal/consul/      两个服务共用的 Consul 客户端
configs/central/      Central 配置示例
configs/node/         Node 配置示例
test/internal/        镜像 internal 包结构的测试
```

Central 和 Node 保持在同一个 Go module 中，但运行时代码和测试按服务隔离。
所有 `*_test.go` 必须放在 `test/` 下，并镜像对应的源代码包路径。
---

## 🔒 许可协议

VideoCluster 使用 Apache 许可证2.0版本，请参阅 LICENSE 文件了解更多。
