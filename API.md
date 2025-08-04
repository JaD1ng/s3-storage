# API 文档

Mock Object Storage Service 提供完整的RESTful API接口，兼容S3协议并扩展了管理功能。

## 基础信息

- **Base URL**: `http://localhost:8080`
- **Content-Type**: `application/json` (管理API) / `*/*` (S3 API)
- **认证**: 当前版本无需认证

## S3兼容API

### 上传对象

**PUT** `/{bucket}/{key}`

上传文件到指定的bucket和key。

#### 请求参数

| 参数 | 类型 | 位置 | 必需 | 描述 |
|------|------|------|------|------|
| bucket | string | path | 是 | 存储桶名称 |
| key | string | path | 是 | 对象键名 |
| Content-Type | string | header | 否 | 文件MIME类型 |

#### 请求体

文件的二进制数据或文本内容。

#### 响应

**成功 (200 OK)**
```json
{
  "message": "Object uploaded successfully",
  "key": "bucket/key",
  "size": 1024
}
```

**错误响应**
```json
{
  "error": "Error message"
}
```

#### 示例

```bash
# 上传文本文件
curl -X PUT "http://localhost:8080/my-bucket/hello.txt" -H "Content-Type: text/plain" -d "Hello, World!"

# 上传二进制文件
curl -X PUT "http://localhost:8080/my-bucket/image.jpg" -H "Content-Type: image/jpeg" --data-binary @image.jpg
```

---

### 下载对象

**GET** `/{bucket}/{key}`

下载指定的对象。

#### 请求参数

| 参数 | 类型 | 位置 | 必需 | 描述 |
|------|------|------|------|------|
| bucket | string | path | 是 | 存储桶名称 |
| key | string | path | 是 | 对象键名 |

#### 响应

**成功 (200 OK)**

返回文件的原始内容，Content-Type根据文件类型设置。

**错误 (404 Not Found)**
```json
{
  "error": "Object not found"
}
```

#### 示例

```bash
# 下载文件
curl -X GET "http://localhost:8080/my-bucket/hello.txt"

# 下载并保存到文件
curl -X GET "http://localhost:8080/my-bucket/hello.txt" -o hello.txt
```

---

### 删除对象

**DELETE** `/{bucket}/{key}`

删除指定的对象。

#### 请求参数

| 参数 | 类型 | 位置 | 必需 | 描述 |
|------|------|------|------|------|
| bucket | string | path | 是 | 存储桶名称 |
| key | string | path | 是 | 对象键名 |

#### 响应

**成功 (204 No Content)**

无响应体。

**错误 (404 Not Found)**
```json
{
  "error": "Object not found"
}
```

#### 示例

```bash
curl -X DELETE "http://localhost:8080/my-bucket/hello.txt"
```

---

### 获取对象元数据

**HEAD** `/{bucket}/{key}`

获取对象的元数据信息，不返回文件内容。

#### 请求参数

| 参数 | 类型 | 位置 | 必需 | 描述 |
|------|------|------|------|------|
| bucket | string | path | 是 | 存储桶名称 |
| key | string | path | 是 | 对象键名 |

#### 响应

**成功 (200 OK)**

响应头包含文件元数据：
- `Content-Type`: 文件MIME类型
- `Content-Length`: 文件大小
- `ETag`: 文件MD5哈希值
- `Last-Modified`: 最后修改时间

#### 示例

```bash
curl -I "http://localhost:8080/my-bucket/hello.txt"
```

---

### 列出对象

**GET** `/{bucket}`

列出指定bucket中的所有对象。

#### 请求参数

| 参数 | 类型 | 位置 | 必需 | 描述 |
|------|------|------|------|------|
| bucket | string | path | 是 | 存储桶名称 |
| prefix | string | query | 否 | 对象键前缀过滤 |
| limit | int | query | 否 | 返回数量限制 (默认100) |
| offset | int | query | 否 | 偏移量 (默认0) |

#### 响应

**成功 (200 OK)**
```json
{
  "objects": [
    {
      "key": "hello.txt",
      "size": 13,
      "last_modified": "2024-01-01T12:00:00Z",
      "etag": "5d41402abc4b2a76b9719d911017c592"
    }
  ],
  "total": 1,
  "truncated": false
}
```

#### 示例

```bash
# 列出所有对象
curl "http://localhost:8080/my-bucket"

# 带前缀过滤
curl "http://localhost:8080/my-bucket?prefix=images/"

# 分页查询
curl "http://localhost:8080/my-bucket?limit=10&offset=20"
```

---

## 管理API

### 列出所有对象

**GET** `/api/v1/objects`

列出系统中的所有对象。

#### 请求参数

| 参数 | 类型 | 位置 | 必需 | 描述 |
|------|------|------|------|------|
| limit | int | query | 否 | 返回数量限制 (默认100) |
| offset | int | query | 否 | 偏移量 (默认0) |

#### 响应

```json
{
  "objects": [
    {
      "id": "uuid",
      "key": "bucket/object.txt",
      "size": 1024,
      "content_type": "text/plain",
      "md5_hash": "hash",
      "created_at": "2024-01-01T12:00:00Z",
      "updated_at": "2024-01-01T12:00:00Z"
    }
  ],
  "total": 1,
  "limit": 100,
  "offset": 0
}
```

---

### 获取对象信息

**GET** `/api/v1/objects/{key}`

获取指定对象的详细信息。

#### 响应

```json
{
  "object": {
    "id": "uuid",
    "key": "bucket/object.txt",
    "size": 1024,
    "content_type": "text/plain",
    "md5_hash": "hash",
    "created_at": "2024-01-01T12:00:00Z",
    "updated_at": "2024-01-01T12:00:00Z"
  },
  "data": "base64_encoded_content"
}
```

---

### API上传对象

**POST** `/api/v1/objects`

通过JSON API上传对象。

#### 请求体

```json
{
  "key": "bucket/object.txt",
  "data": "base64_encoded_content",
  "content_type": "text/plain"
}
```

#### 响应

```json
{
  "message": "Object uploaded successfully",
  "object": {
    "id": "uuid",
    "key": "bucket/object.txt",
    "size": 1024
  }
}
```

---

### API删除对象

**DELETE** `/api/v1/objects/{key}`

通过API删除对象。

#### 响应

```json
{
  "message": "Object deleted successfully"
}
```

---

### 获取统计信息

**GET** `/api/v1/stats`

获取系统统计信息。

#### 响应

```json
{
  "total_objects": 100,
  "total_size": 1048576,
  "storage_nodes": {
    "stg1": {
      "status": "healthy",
      "objects": 100,
      "size": 1048576
    },
    "stg2": {
      "status": "healthy", 
      "objects": 100,
      "size": 1048576
    },
    "stg3": {
      "status": "healthy",
      "objects": 100,
      "size": 1048576
    }
  },
  "queue_status": {
    "pending_tasks": 5,
    "completed_tasks": 95,
    "failed_tasks": 0
  }
}
```

---

### 搜索对象

**GET** `/api/v1/search`

搜索对象。

#### 请求参数

| 参数 | 类型 | 位置 | 必需 | 描述 |
|------|------|------|------|------|
| q | string | query | 是 | 搜索关键词 |
| limit | int | query | 否 | 返回数量限制 (默认50) |
| offset | int | query | 否 | 偏移量 (默认0) |

#### 响应

```json
{
  "results": [
    {
      "key": "bucket/matching-object.txt",
      "size": 1024,
      "content_type": "text/plain",
      "score": 0.95
    }
  ],
  "total": 1,
  "query": "matching"
}
```

---

## 系统API

### 健康检查

**GET** `/health`

检查服务健康状态。

#### 响应

```json
{
  "status": "healthy",
  "timestamp": "2024-01-01T12:00:00Z",
  "version": "1.0.0",
  "components": {
    "database": "healthy",
    "storage": "healthy",
    "queue": "healthy"
  }
}
```

---

## 错误代码

| HTTP状态码 | 描述 |
|------------|------|
| 200 | 请求成功 |
| 201 | 创建成功 |
| 204 | 删除成功 |
| 400 | 请求参数错误 |
| 404 | 资源不存在 |
| 409 | 资源冲突 |
| 500 | 服务器内部错误 |

## 限制说明

- 单个文件最大大小：无限制（受磁盘空间限制）
- 并发请求数：无限制（受系统资源限制）
- 对象键长度：最大1024字符
- Bucket名称：支持字母、数字、连字符

## 第三方集成

当请求的对象在本地存储中不存在时，系统会自动尝试从配置的第三方服务获取数据：

1. 检查本地存储
2. 如果不存在，调用第三方API
3. 将获取的数据保存到本地存储
4. 返回数据给客户端

第三方服务需要实现标准的HTTP GET接口。