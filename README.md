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
- gRPC control plane for node management
- HTTP streaming with Range request support
- Centralized aggregation and metadata indexing
- Support for future horizontal scaling

---

## 🏗 Architecture

- **Central Server**
    - Video aggregation
    - Authentication
    - Authorization
    - Node registry
    - Playlist service

- **Node Server**
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

## 📂 Project Structure


---

## 🔒 License

VideoCluster uses Apache license version 2.0. Please refer to the license file for more information.

