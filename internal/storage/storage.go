package storage

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"mock-storage/internal/types"
)

// FileStorageNode 基于文件系统的存储节点实现
type FileStorageNode struct {
	nodeID   string
	basePath string
}

// NewFileStorageNode 创建新的文件存储节点
func NewFileStorageNode(nodeID, basePath string) (*FileStorageNode, error) {
	// 确保目录存在
	err := os.MkdirAll(basePath, 0755)
	if err != nil {
		return nil, fmt.Errorf("failed to create storage directory %s: %v", basePath, err)
	}

	return &FileStorageNode{
		nodeID:   nodeID,
		basePath: basePath,
	}, nil
}

// GetNodeID 获取节点ID
func (fs *FileStorageNode) GetNodeID() string {
	return fs.nodeID
}

// Write 写入文件对象到存储节点
func (fs *FileStorageNode) Write(obj *types.FileObject) error {
	// 生成文件路径
	filePath := fs.getFilePath(obj.Key)

	// 确保目录存在
	dir := filepath.Dir(filePath)
	err := os.MkdirAll(dir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create directory %s: %v", dir, err)
	}

	// 写入文件数据
	err = os.WriteFile(filePath, obj.Data, 0644)
	if err != nil {
		return fmt.Errorf("failed to write file %s: %v", filePath, err)
	}

	// 验证MD5
	if obj.MD5Hash != "" {
		calculatedHash := calculateMD5(obj.Data)
		if calculatedHash != obj.MD5Hash {
			// 删除写入的文件
			os.Remove(filePath)
			return fmt.Errorf("MD5 hash mismatch: expected %s, got %s", obj.MD5Hash, calculatedHash)
		}
	}

	fmt.Printf("[%s] Successfully wrote file: %s (size: %d bytes)\n", fs.nodeID, obj.Key, len(obj.Data))
	return nil
}

// Read 从存储节点读取文件对象
func (fs *FileStorageNode) Read(key string) (*types.FileObject, error) {
	filePath := fs.getFilePath(key)

	// 检查文件是否存在
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("file not found: %s", key)
	}

	// 读取文件数据
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read file %s: %v", filePath, err)
	}

	// 获取文件信息
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file info %s: %v", filePath, err)
	}

	obj := &types.FileObject{
		Key:       key,
		Size:      fileInfo.Size(),
		Data:      data,
		MD5Hash:   calculateMD5(data),
		CreatedAt: fileInfo.ModTime(),
	}

	return obj, nil
}

// Delete 从存储节点删除文件
func (fs *FileStorageNode) Delete(key string) error {
	filePath := fs.getFilePath(key)

	err := os.Remove(filePath)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file %s: %v", filePath, err)
	}

	fmt.Printf("[%s] Successfully deleted file: %s\n", fs.nodeID, key)
	return nil
}

// getFilePath 根据key生成文件路径
func (fs *FileStorageNode) getFilePath(key string) string {
	// 直接使用key作为相对路径，保持bucket/object的层次结构
	return filepath.Join(fs.basePath, key)
}

// calculateMD5 计算数据的MD5哈希
func calculateMD5(data []byte) string {
	hash := md5.Sum(data)
	return hex.EncodeToString(hash[:])
}
