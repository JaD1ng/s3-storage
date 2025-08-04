package s3

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
)

// ListObjects 处理LIST对象请求
func (h *Handler) ListObjects(c *gin.Context) {
	bucket := c.Param("bucket")

	// 获取查询参数
	limitStr := c.DefaultQuery("max-keys", "1000")
	offsetStr := c.DefaultQuery("marker", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 1000
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		offset = 0
	}

	// 获取元数据列表
	metadataList, err := h.service.ListMetadata(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to list objects",
		})
		return
	}

	// 过滤属于指定bucket的对象
	var filteredObjects []gin.H
	bucketPrefix := bucket + "/"

	for _, metadata := range metadataList {
		if strings.HasPrefix(metadata.Key, bucketPrefix) {
			// 移除bucket前缀，只返回对象key
			objectKey := h.extractObjectKey(metadata.Key, bucketPrefix)
			filteredObjects = append(filteredObjects, gin.H{
				"Key":          objectKey,
				"Size":         metadata.Size,
				"LastModified": metadata.UpdatedAt.Format("2006-01-02T15:04:05.000Z"),
				"ETag":         `"` + metadata.MD5Hash + `"`,
				"StorageClass": "STANDARD",
			})
		}
	}

	// 构建S3兼容的响应
	response := gin.H{
		"Name":           bucket,
		"Prefix":         "",
		"Marker":         offsetStr,
		"MaxKeys":        limit,
		"IsTruncated":    len(filteredObjects) == limit,
		"Contents":       filteredObjects,
		"CommonPrefixes": []string{},
	}

	c.XML(http.StatusOK, response)
}

// HeadObject 处理HEAD对象请求
func (h *Handler) HeadObject(c *gin.Context) {
	bucket := c.Param("bucket")
	key := c.Param("key")

	// 构建对象key（包含bucket前缀）
	objectKey := h.buildObjectKey(bucket, key)

	// 从元数据服务获取文件信息
	metadata, err := h.service.GetMetadata(objectKey)
	if err != nil {
		c.Status(http.StatusNotFound)
		return
	}

	// 设置响应头
	c.Header("Content-Type", metadata.ContentType)
	c.Header("Content-Length", strconv.FormatInt(metadata.Size, 10))
	c.Header("ETag", `"`+metadata.MD5Hash+`"`)
	c.Header("Last-Modified", metadata.UpdatedAt.Format(http.TimeFormat))

	c.Status(http.StatusOK)
}

// ListObjectsAPI 处理API LIST对象请求
func (h *Handler) ListObjectsAPI(c *gin.Context) {
	limitStr := c.DefaultQuery("limit", "100")
	offsetStr := c.DefaultQuery("offset", "0")

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 100
	}

	offset, err := strconv.Atoi(offsetStr)
	if err != nil {
		offset = 0
	}

	metadataList, err := h.service.ListMetadata(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"objects": metadataList,
		"total":   len(metadataList),
		"limit":   limit,
		"offset":  offset,
	})
}

// GetStatsAPI 处理获取统计信息请求
func (h *Handler) GetStatsAPI(c *gin.Context) {
	stats, err := h.service.GetStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// SearchObjectsAPI 处理搜索对象请求
func (h *Handler) SearchObjectsAPI(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Query parameter 'q' is required"})
		return
	}

	limitStr := c.DefaultQuery("limit", "50")
	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		limit = 50
	}

	results, err := h.service.SearchMetadata(query, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"query":   query,
		"results": results,
		"total":   len(results),
		"limit":   limit,
	})
}