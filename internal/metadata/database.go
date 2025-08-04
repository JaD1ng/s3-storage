package metadata

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"mock-storage/internal/types"

	_ "github.com/mattn/go-sqlite3"
)

// DatabaseManager 数据库管理器
type DatabaseManager struct {
	db *sql.DB
}

// NewDatabaseManager 创建数据库管理器
func NewDatabaseManager(driver, dsn string) (*DatabaseManager, error) {
	db, err := sql.Open(driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %v", err)
	}

	// 测试连接
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to ping database: %v", err)
	}

	manager := &DatabaseManager{db: db}

	// 初始化表结构
	err = manager.initTables()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize tables: %v", err)
	}

	fmt.Println("[DB] Database connected and initialized successfully")
	return manager, nil
}

// initTables 初始化数据库表
func (dm *DatabaseManager) initTables() error {
	createTableSQL := `
	CREATE TABLE IF NOT EXISTS metadata (
		id TEXT PRIMARY KEY,
		key TEXT UNIQUE NOT NULL,
		size INTEGER NOT NULL,
		content_type TEXT NOT NULL,
		md5_hash TEXT NOT NULL,
		storage_nodes TEXT NOT NULL, -- JSON array
		created_at DATETIME NOT NULL,
		updated_at DATETIME NOT NULL
	);
	
	CREATE INDEX IF NOT EXISTS idx_metadata_key ON metadata(key);
	CREATE INDEX IF NOT EXISTS idx_metadata_created_at ON metadata(created_at);
	CREATE INDEX IF NOT EXISTS idx_metadata_size ON metadata(size);
	`

	_, err := dm.db.Exec(createTableSQL)
	return err
}

// SaveMetadata 保存元数据到数据库
func (dm *DatabaseManager) SaveMetadata(entry *types.MetadataEntry) error {
	// 将storage_nodes转换为JSON字符串
	storageNodesJSON, err := json.Marshal(entry.StorageNodes)
	if err != nil {
		return fmt.Errorf("failed to marshal storage nodes: %v", err)
	}

	insertSQL := `
	INSERT OR REPLACE INTO metadata 
	(id, key, size, content_type, md5_hash, storage_nodes, created_at, updated_at)
	VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`

	_, err = dm.db.Exec(insertSQL,
		entry.ID,
		entry.Key,
		entry.Size,
		entry.ContentType,
		entry.MD5Hash,
		string(storageNodesJSON),
		entry.CreatedAt,
		entry.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to insert metadata: %v", err)
	}

	fmt.Printf("[DB] Saved metadata for key: %s\n", entry.Key)
	return nil
}

// GetMetadata 从数据库获取元数据
func (dm *DatabaseManager) GetMetadata(key string) (*types.MetadataEntry, error) {
	querySQL := `
	SELECT id, key, size, content_type, md5_hash, storage_nodes, created_at, updated_at
	FROM metadata WHERE key = ?
	`

	row := dm.db.QueryRow(querySQL, key)

	var entry types.MetadataEntry
	var storageNodesJSON string
	var createdAt, updatedAt string

	err := row.Scan(
		&entry.ID,
		&entry.Key,
		&entry.Size,
		&entry.ContentType,
		&entry.MD5Hash,
		&storageNodesJSON,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("metadata not found for key: %s", key)
		}
		return nil, fmt.Errorf("failed to query metadata: %v", err)
	}

	// 解析JSON字符串为storage_nodes数组
	err = json.Unmarshal([]byte(storageNodesJSON), &entry.StorageNodes)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal storage nodes: %v", err)
	}

	// 解析时间
	entry.CreatedAt, err = time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse created_at: %v", err)
	}

	entry.UpdatedAt, err = time.Parse(time.RFC3339, updatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse updated_at: %v", err)
	}

	return &entry, nil
}

// DeleteMetadata 从数据库删除元数据
func (dm *DatabaseManager) DeleteMetadata(key string) error {
	deleteSQL := `DELETE FROM metadata WHERE key = ?`

	result, err := dm.db.Exec(deleteSQL, key)
	if err != nil {
		return fmt.Errorf("failed to delete metadata: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("metadata not found for key: %s", key)
	}

	fmt.Printf("[DB] Deleted metadata for key: %s\n", key)
	return nil
}

// ListMetadata 列出元数据（分页）
func (dm *DatabaseManager) ListMetadata(limit, offset int) ([]*types.MetadataEntry, error) {
	querySQL := `
	SELECT id, key, size, content_type, md5_hash, storage_nodes, created_at, updated_at
	FROM metadata 
	ORDER BY created_at DESC
	LIMIT ? OFFSET ?
	`

	rows, err := dm.db.Query(querySQL, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to query metadata list: %v", err)
	}
	defer rows.Close()

	var entries []*types.MetadataEntry

	for rows.Next() {
		var entry types.MetadataEntry
		var storageNodesJSON string
		var createdAt, updatedAt string

		err := rows.Scan(
			&entry.ID,
			&entry.Key,
			&entry.Size,
			&entry.ContentType,
			&entry.MD5Hash,
			&storageNodesJSON,
			&createdAt,
			&updatedAt,
		)

		if err != nil {
			return nil, fmt.Errorf("failed to scan metadata row: %v", err)
		}

		// 解析JSON和时间
		err = json.Unmarshal([]byte(storageNodesJSON), &entry.StorageNodes)
		if err != nil {
			continue // 跳过损坏的记录
		}

		entry.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		entry.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

		entries = append(entries, &entry)
	}

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %v", err)
	}

	return entries, nil
}

// UpdateMetadata 更新元数据
func (dm *DatabaseManager) UpdateMetadata(entry *types.MetadataEntry) error {
	storageNodesJSON, err := json.Marshal(entry.StorageNodes)
	if err != nil {
		return fmt.Errorf("failed to marshal storage nodes: %v", err)
	}

	updateSQL := `
	UPDATE metadata 
	SET size = ?, content_type = ?, md5_hash = ?, storage_nodes = ?, updated_at = ?
	WHERE key = ?
	`

	result, err := dm.db.Exec(updateSQL,
		entry.Size,
		entry.ContentType,
		entry.MD5Hash,
		string(storageNodesJSON),
		entry.UpdatedAt,
		entry.Key,
	)

	if err != nil {
		return fmt.Errorf("failed to update metadata: %v", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %v", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("metadata not found for key: %s", entry.Key)
	}

	return nil
}

// GetStats 获取统计信息
func (dm *DatabaseManager) GetStats() (map[string]any, error) {
	stats := make(map[string]any)

	// 总文件数
	var totalFiles int64
	err := dm.db.QueryRow("SELECT COUNT(*) FROM metadata").Scan(&totalFiles)
	if err != nil {
		return nil, err
	}
	stats["total_files"] = totalFiles

	// 总大小
	var totalSize sql.NullInt64
	err = dm.db.QueryRow("SELECT SUM(size) FROM metadata").Scan(&totalSize)
	if err != nil {
		return nil, err
	}
	if totalSize.Valid {
		stats["total_size"] = totalSize.Int64
	} else {
		stats["total_size"] = 0
	}

	// 平均大小
	if totalFiles > 0 {
		stats["average_size"] = totalSize.Int64 / totalFiles
	} else {
		stats["average_size"] = 0
	}

	// 按内容类型统计
	contentTypeStats := make(map[string]int)
	rows, err := dm.db.Query("SELECT content_type, COUNT(*) FROM metadata GROUP BY content_type")
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var contentType string
			var count int
			if rows.Scan(&contentType, &count) == nil {
				contentTypeStats[contentType] = count
			}
		}
	}
	stats["content_types"] = contentTypeStats

	return stats, nil
}

// SearchMetadata 搜索元数据
func (dm *DatabaseManager) SearchMetadata(query string, limit int) ([]*types.MetadataEntry, error) {
	searchSQL := `
	SELECT id, key, size, content_type, md5_hash, storage_nodes, created_at, updated_at
	FROM metadata 
	WHERE key LIKE ? OR content_type LIKE ?
	ORDER BY created_at DESC
	LIMIT ?
	`

	searchPattern := "%" + strings.ToLower(query) + "%"
	rows, err := dm.db.Query(searchSQL, searchPattern, searchPattern, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to search metadata: %v", err)
	}
	defer rows.Close()

	var entries []*types.MetadataEntry

	for rows.Next() {
		var entry types.MetadataEntry
		var storageNodesJSON string
		var createdAt, updatedAt string

		err := rows.Scan(
			&entry.ID,
			&entry.Key,
			&entry.Size,
			&entry.ContentType,
			&entry.MD5Hash,
			&storageNodesJSON,
			&createdAt,
			&updatedAt,
		)

		if err != nil {
			continue
		}

		err = json.Unmarshal([]byte(storageNodesJSON), &entry.StorageNodes)
		if err != nil {
			continue
		}

		entry.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		entry.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

		entries = append(entries, &entry)
	}

	return entries, nil
}

// Close 关闭数据库连接
func (dm *DatabaseManager) Close() error {
	if dm.db != nil {
		return dm.db.Close()
	}
	return nil
}
