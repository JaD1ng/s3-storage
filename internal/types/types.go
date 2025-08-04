package types

import (
	"time"
)

// FileObject 表示上传的文件对象
type FileObject struct {
	ID          string    `json:"id"`
	Key         string    `json:"key"`
	Size        int64     `json:"size"`
	ContentType string    `json:"content_type"`
	MD5Hash     string    `json:"md5_hash"`
	Data        []byte    `json:"-"` // 文件数据，不序列化到JSON
	CreatedAt   time.Time `json:"created_at"`
}

// MetadataEntry 元数据条目
type MetadataEntry struct {
	ID           string    `json:"id" db:"id"`
	Key          string    `json:"key" db:"key"`
	Size         int64     `json:"size" db:"size"`
	ContentType  string    `json:"content_type" db:"content_type"`
	MD5Hash      string    `json:"md5_hash" db:"md5_hash"`
	StorageNodes []string  `json:"storage_nodes" db:"storage_nodes"` // 存储节点列表
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}

// StorageNode 存储节点接口
type StorageNode interface {
	Write(obj *FileObject) error
	Read(key string) (*FileObject, error)
	Delete(key string) error
	GetNodeID() string
}

// UploadRequest 上传请求
type UploadRequest struct {
	Key         string `json:"key"`
	ContentType string `json:"content_type"`
	Data        []byte `json:"data"`
}

// UploadResponse 上传响应
type UploadResponse struct {
	Success  bool   `json:"success"`
	ObjectID string `json:"object_id,omitempty"`
	Key      string `json:"key,omitempty"`
	Message  string `json:"message,omitempty"`
	Size     int64  `json:"size,omitempty"`
	MD5Hash  string `json:"md5_hash,omitempty"`
}

// TaskMessage 队列任务消息
type TaskMessage struct {
	Type      string         `json:"type"`
	ObjectID  string         `json:"object_id"`
	Data      map[string]any `json:"data"`
	CreatedAt time.Time      `json:"created_at"`
}

// TaskProcessor 任务处理器接口
type TaskProcessor interface {
	Process(task *TaskMessage) error
}
