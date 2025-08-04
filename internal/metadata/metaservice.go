package metadata

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"mock-storage/internal/types"
)

// MetaService 元数据服务
type MetaService struct {
	db *DatabaseManager
}

// NewMetaService 创建元数据服务
func NewMetaService(db *DatabaseManager) *MetaService {
	return &MetaService{
		db: db,
	}
}

// SaveMetadata 保存元数据
func (ms *MetaService) SaveMetadata(obj *types.FileObject, storageNodes []string) error {
	entry := &types.MetadataEntry{
		ID:           obj.ID,
		Key:          obj.Key,
		Size:         obj.Size,
		ContentType:  obj.ContentType,
		MD5Hash:      obj.MD5Hash,
		StorageNodes: storageNodes,
		CreatedAt:    obj.CreatedAt,
		UpdatedAt:    time.Now(),
	}

	err := ms.db.SaveMetadata(entry)
	if err != nil {
		return fmt.Errorf("failed to save metadata: %v", err)
	}

	fmt.Printf("[META] Successfully saved metadata for key: %s\n", obj.Key)
	return nil
}

// GetMetadata 获取元数据
func (ms *MetaService) GetMetadata(key string) (*types.MetadataEntry, error) {
	entry, err := ms.db.GetMetadata(key)
	if err != nil {
		return nil, fmt.Errorf("failed to get metadata: %v", err)
	}

	return entry, nil
}

// DeleteMetadata 删除元数据
func (ms *MetaService) DeleteMetadata(key string) error {
	err := ms.db.DeleteMetadata(key)
	if err != nil {
		return fmt.Errorf("failed to delete metadata: %v", err)
	}

	fmt.Printf("[META] Successfully deleted metadata for key: %s\n", key)
	return nil
}

// ListMetadata 列出元数据（分页）
func (ms *MetaService) ListMetadata(limit, offset int) ([]*types.MetadataEntry, error) {
	entries, err := ms.db.ListMetadata(limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list metadata: %v", err)
	}

	return entries, nil
}

// UpdateMetadata 更新元数据
func (ms *MetaService) UpdateMetadata(key string, updates map[string]any) error {
	// 首先获取现有元数据
	entry, err := ms.GetMetadata(key)
	if err != nil {
		return err
	}

	// 更新字段
	if contentType, ok := updates["content_type"].(string); ok {
		entry.ContentType = contentType
	}
	if storageNodes, ok := updates["storage_nodes"].([]string); ok {
		entry.StorageNodes = storageNodes
	}

	entry.UpdatedAt = time.Now()

	err = ms.db.UpdateMetadata(entry)
	if err != nil {
		return fmt.Errorf("failed to update metadata: %v", err)
	}

	fmt.Printf("[META] Successfully updated metadata for key: %s\n", key)
	return nil
}

// GetStats 获取统计信息
func (ms *MetaService) GetStats() (map[string]any, error) {
	stats, err := ms.db.GetStats()
	if err != nil {
		return nil, fmt.Errorf("failed to get stats: %v", err)
	}

	return stats, nil
}

// SearchMetadata 搜索元数据
func (ms *MetaService) SearchMetadata(query string, limit int) ([]*types.MetadataEntry, error) {
	// 简单的关键字搜索实现
	entries, err := ms.db.SearchMetadata(query, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search metadata: %v", err)
	}

	return entries, nil
}

// ValidateMetadata 验证元数据
func (ms *MetaService) ValidateMetadata(entry *types.MetadataEntry) error {
	if entry.Key == "" {
		return fmt.Errorf("key cannot be empty")
	}

	if entry.Size < 0 {
		return fmt.Errorf("size cannot be negative")
	}

	if len(entry.StorageNodes) == 0 {
		return fmt.Errorf("storage nodes cannot be empty")
	}

	// 验证MD5格式
	if entry.MD5Hash != "" && len(entry.MD5Hash) != 32 {
		return fmt.Errorf("invalid MD5 hash format")
	}

	return nil
}

// ExportMetadata 导出元数据到JSON
func (ms *MetaService) ExportMetadata(keys []string) ([]byte, error) {
	var entries []*types.MetadataEntry

	if len(keys) == 0 {
		// 导出所有元数据
		allEntries, err := ms.ListMetadata(1000, 0) // 限制1000条
		if err != nil {
			return nil, err
		}
		entries = allEntries
	} else {
		// 导出指定的元数据
		for _, key := range keys {
			entry, err := ms.GetMetadata(key)
			if err != nil {
				continue // 跳过不存在的key
			}
			entries = append(entries, entry)
		}
	}

	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %v", err)
	}

	return data, nil
}

// ImportMetadata 从JSON导入元数据
func (ms *MetaService) ImportMetadata(data []byte) error {
	var entries []*types.MetadataEntry

	err := json.Unmarshal(data, &entries)
	if err != nil {
		return fmt.Errorf("failed to unmarshal metadata: %v", err)
	}

	successCount := 0
	for _, entry := range entries {
		err := ms.ValidateMetadata(entry)
		if err != nil {
			fmt.Printf("Warning: invalid metadata entry %s: %v\n", entry.Key, err)
			continue
		}

		err = ms.db.SaveMetadata(entry)
		if err != nil {
			fmt.Printf("Warning: failed to import metadata %s: %v\n", entry.Key, err)
			continue
		}

		successCount++
	}

	fmt.Printf("[META] Successfully imported %d out of %d metadata entries\n", successCount, len(entries))
	return nil
}

// GetMetadataByPattern 根据模式获取元数据
func (ms *MetaService) GetMetadataByPattern(pattern string) ([]*types.MetadataEntry, error) {
	// 简单的通配符匹配实现
	allEntries, err := ms.ListMetadata(1000, 0)
	if err != nil {
		return nil, err
	}

	var matchedEntries []*types.MetadataEntry
	for _, entry := range allEntries {
		if matchPattern(entry.Key, pattern) {
			matchedEntries = append(matchedEntries, entry)
		}
	}

	return matchedEntries, nil
}

// matchPattern 简单的通配符匹配
func matchPattern(text, pattern string) bool {
	if pattern == "*" {
		return true
	}

	if strings.Contains(pattern, "*") {
		// 简单的前缀/后缀匹配
		if strings.HasPrefix(pattern, "*") && strings.HasSuffix(pattern, "*") {
			// *something*
			middle := pattern[1 : len(pattern)-1]
			return strings.Contains(text, middle)
		} else if strings.HasPrefix(pattern, "*") {
			// *suffix
			suffix := pattern[1:]
			return strings.HasSuffix(text, suffix)
		} else if strings.HasSuffix(pattern, "*") {
			// prefix*
			prefix := pattern[:len(pattern)-1]
			return strings.HasPrefix(text, prefix)
		}
	}

	return text == pattern
}
