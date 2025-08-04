# Mock Object Storage Service

一个轻量级的对象存储服务，提供S3兼容的API接口，支持文件上传、下载、删除等基本操作，并具备第三方数据获取和异步任务处理能力。

## 🚀 特性

- **S3兼容API**: 提供标准的对象存储接口
- **多节点存储**: 支持多个存储节点的数据冗余
- **第三方集成**: 自动从第三方服务获取不存在的对象
- **异步任务处理**: 基于队列的异步任务处理机制
- **元数据管理**: SQLite数据库存储文件元数据
- **健康检查**: 内置服务健康监控
- **管理API**: 提供丰富的管理和统计接口

## 📋 系统要求

- Go 1.21+
- SQLite3
- 磁盘空间（用于存储文件数据）

## 🛠️ 安装与构建

### 克隆项目
```bash
git clone <repository-url>
cd mock-storage
```

### 安装依赖
```bash
make deps
```

### 构建项目
```bash
make build
```

### 运行服务
```bash
make run
```

## ⚙️ 配置

项目使用 `config.json` 文件进行配置：

```json
{
  "server": {
    "port": "8080",
    "host": "localhost"
  },
  "storage": {
    "data_dir": "./data",
    "nodes": [
      {
        "id": "stg1",
        "path": "./data/stg1"
      },
      {
        "id": "stg2", 
        "path": "./data/stg2"
      },
      {
        "id": "stg3",
        "path": "./data/stg3"
      }
    ]
  },
  "database": {
    "driver": "sqlite3",
    "dsn": "./data/metadata.db"
  },
  "queue": {
    "size": 1000
  }
}
```

## 📡 API 接口

### S3兼容接口

| 方法 | 路径 | 描述 |
|------|------|------|
| PUT | `/{bucket}/{key}` | 上传对象 |
| GET | `/{bucket}/{key}` | 下载对象 |
| DELETE | `/{bucket}/{key}` | 删除对象 |
| HEAD | `/{bucket}/{key}` | 获取对象元数据 |
| GET | `/{bucket}` | 列出bucket中的对象 |

### 管理API

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | `/api/v1/objects` | 列出所有对象 |
| GET | `/api/v1/objects/{key}` | 获取指定对象信息 |
| POST | `/api/v1/objects` | 通过API上传对象 |
| DELETE | `/api/v1/objects/{key}` | 通过API删除对象 |
| GET | `/api/v1/stats` | 获取系统统计信息 |
| GET | `/api/v1/search?q={query}` | 搜索对象 |

### 系统接口

| 方法 | 路径 | 描述 |
|------|------|------|
| GET | `/health` | 健康检查 |

## 🔧 使用示例

### 上传文件
```bash
curl -X PUT "http://localhost:8080/my-bucket/test.txt" -H "Content-Type: text/plain" -d "Hello, World!"
```

### 下载文件
```bash
curl -X GET "http://localhost:8080/my-bucket/test.txt"
```

### 删除文件
```bash
curl -X DELETE "http://localhost:8080/my-bucket/test.txt"
```

### 列出对象
```bash
curl -X GET "http://localhost:8080/my-bucket"
```

### 获取统计信息
```bash
curl -X GET "http://localhost:8080/api/v1/stats"
```

## 🏗️ 架构设计

### 核心组件

- **Storage Manager**: 管理多个存储节点
- **Metadata Service**: 处理文件元数据的存储和查询
- **Queue Manager**: 异步任务队列管理
- **Third Party Service**: 第三方数据获取服务
- **S3 Handler**: S3兼容的HTTP接口处理

### 数据流程

1. **上传流程**:
   - 接收HTTP请求
   - 写入多个存储节点
   - 保存元数据到数据库
   - 异步处理上传完成任务

2. **下载流程**:
   - 查询元数据
   - 从存储节点读取数据
   - 如果不存在，从第三方服务获取
   - 返回文件内容

3. **删除流程**:
   - 删除元数据记录
   - 异步删除存储节点中的文件

## 📁 项目结构

```
mock-storage/
├── cmd/server/          # 应用入口
├── internal/
│   ├── config/          # 配置管理
│   ├── handler/s3/      # S3接口处理器
│   ├── metadata/        # 元数据服务
│   ├── queue/           # 队列管理
│   ├── service/         # 核心服务
│   ├── storage/         # 存储管理
│   ├── types/           # 数据类型定义
│   └── utils/           # 工具函数
├── data/                # 数据目录
│   ├── metadata.db      # SQLite数据库
│   ├── stg1/            # 存储节点1
│   ├── stg2/            # 存储节点2
│   └── stg3/            # 存储节点3
├── config.json          # 配置文件
└── Makefile            # 构建脚本
```

## 🔨 开发

### 可用的Make命令

```bash
make build      # 构建项目
make run        # 运行服务
make test       # 运行测试
make clean      # 清理构建文件
make deps       # 安装依赖
make fmt        # 格式化代码
make help       # 显示帮助信息
```

### 多平台构建

```bash
make build-all     # 构建所有平台版本
make build-linux   # 构建Linux版本
make build-windows # 构建Windows版本
make build-darwin  # 构建MacOS版本
```

## 🚦 健康检查

服务启动后，可以通过以下方式检查服务状态：

```bash
curl http://localhost:8080/health
```

## 📊 监控与统计

通过管理API可以获取系统运行统计信息：

```bash
curl http://localhost:8080/api/v1/stats
```

返回信息包括：
- 总对象数量
- 存储使用情况
- 系统运行状态

## 📝 TODO

- [ ] **Prometheus 监控**: 集成 Prometheus 监控系统
  - 系统性能指标：CPU、内存、磁盘使用率
  - 业务指标：请求QPS、响应时间、错误率
  - 存储指标：文件上传/下载速度、存储空间使用情况
  - 自定义指标：并发连接数、队列长度、缓存命中率
  - Grafana 仪表板：可视化监控数据和告警

- [ ] **Superset 日志分析**: 将日志数据传入 Apache Superset
  - 结构化日志输出：JSON 格式的访问日志和错误日志
  - 日志收集管道：使用 Fluentd/Logstash 收集和转换日志
  - 数据存储：将日志数据存储到时序数据库（如 InfluxDB）
  - Superset 集成：创建日志分析仪表板和报表
  - 实时分析：用户行为分析、性能趋势、异常检测

- [ ] **事务支持**: 实现数据库事务机制，确保元数据操作的原子性
  - 上传操作：元数据写入和文件存储要么全部成功，要么全部回滚
  - 删除操作：元数据删除和文件清理的原子性保证
  - 更新操作：文件替换和元数据更新的一致性

- [ ] **分布式锁**: 实现分布式锁机制，防止并发操作冲突
  - 对象级别的锁定机制
  - 防止同一对象的并发写入操作
  - 超时和死锁检测

- [ ] **两阶段提交**: 对于多节点存储操作实现两阶段提交协议
  - 准备阶段：所有存储节点预留空间和资源
  - 提交阶段：确认所有节点操作成功后统一提交
  - 回滚机制：任一节点失败时的完整回滚

- [ ] **操作日志**: 实现操作日志（WAL）机制
  - 记录所有关键操作的日志
  - 支持操作重放和恢复
  - 崩溃后的数据一致性恢复

- [ ] **一致性检查**: 定期检查数据一致性
  - 元数据与实际文件的一致性验证
  - 多节点间数据同步状态检查
  - 自动修复不一致的数据
