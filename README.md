# VidCluster
[![license](https://img.shields.io/github/license/apache/incubator-seata-go.svg)](https://www.apache.org/licenses/LICENSE-2.0.html)

**Distributed Video Aggregation and Streaming System**

[中文文档 🇨🇳](./README_CN.md)

---

## 📦 Overview

VidCluster is a distributed video management system that supports multi-node video storage, aggregation, and streaming services.

- 🚀 Built with Go and Kratos framework
- 🛰️ Designed for scalability and extensibility
- 🔐 Supports authentication and permission management (future feature)

---

## 🎯 Features

- Distributed multi-node video service
- gRPC control plane for Media Node management
- HTTP streaming with Range request support
- Centralized aggregation and metadata indexing
- Local video directory scanning with metadata stored in BadgerDB
- Fast video fingerprints based on file size and 1 MiB samples from the beginning, middle, and end
- Existing videos skip thumbnail generation and database writes
- Windows Shell thumbnails on Windows and ffmpeg thumbnails on Linux/macOS
- Support for future horizontal scaling

> Media Node scanning, metadata queries, and video streaming are available. Media Node
> discovery and aggregation in the Central service are still under development.
> Authentication, authorization, and playlists are planned features.

---

## 🔍 Video Scanning

The Media Node service scans configured directories for `.mp4`, `.mkv`, `.avi`,
`.mov`, and `.flv` files and processes them as follows:

1. Read the file size and up to 1 MiB from the beginning, middle, and end.
2. Generate a fast BLAKE3 fingerprint from those samples and use it as the video ID.
3. Skip the video when that ID already exists in BadgerDB.
4. Otherwise, generate a platform-specific thumbnail and store the metadata.

The fingerprint reads at most about 3 MiB per video, making it suitable for
scanning large media collections. It is not a full-file hash: a change outside
the sampled regions that preserves the file size may not change the fingerprint.

### Platform differences

| Platform | Thumbnail implementation | Runtime requirement |
| --- | --- | --- |
| Windows | Windows Shell `IShellItemImageFactory` | Uses installed system codecs; ffmpeg is not required |
| Linux / macOS | Frame extraction with ffmpeg | `ffmpeg` must be available in `PATH` |

If Windows cannot decode a video format, the metadata is still stored but its
thumbnail remains empty.

---

## 🏗 Architecture

- **Central Server**
    - Video aggregation
    - Authentication
    - Authorization
    - Media Node registry
    - Playlist service

- **Media Node**
    - Local video metadata service
    - HTTP streaming server

---

## ⚙ Technology Stack

- Golang
- Kratos
- gRPC
- HTTP (Range Streaming)
- Consul
- ffmpeg
- BadgerDB

---

## ⚠️ Data Compatibility

Earlier versions used a full-file BLAKE3 hash as the video ID. The current
version uses a sampled fingerprint, and existing IDs are not migrated
automatically. Development installations can remove their old `*.data`
directory; production data should be migrated before upgrading.

---

## 📂 Project Structure

```text
api/                  Protobuf APIs for Central and Node
cmd/central/          Central entry point and Wire assembly
cmd/node/             Media Node entry point and Wire assembly
internal/central/     Central data, service, and transport layers
internal/node/        Media Node business, data, media, service, and transport layers
internal/conf/        Configuration definitions shared by both services
internal/consul/      Consul client shared by both services
configs/central/      Central configuration example
configs/node/         Media Node configuration example
test/internal/        Tests mirroring the internal package structure
```

Central and Media Node remain in one Go module, while runtime code and tests are
isolated by service. Every `*_test.go` file belongs under `test/`, mirroring its
corresponding source package path.
---

## 🔒 License

VideoCluster uses Apache license version 2.0. Please refer to the license file for more information.
