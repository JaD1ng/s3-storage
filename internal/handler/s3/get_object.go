package s3

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// GetObject 处理GET对象请求
func (h *Handler) GetObject(c *gin.Context) {
	bucket := c.Param("bucket")
	key := c.Param("key")

	// 构建对象key（包含bucket前缀）
	objectKey := h.buildObjectKey(bucket, key)

	// 从元数据服务获取文件信息
	metadata, err := h.service.GetMetadata(objectKey)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": fmt.Sprintf("Object not found: %v", err),
		})
		return
	}

	// 如果元数据中没有存储节点信息，尝试从第三方获取并上传
	if len(metadata.StorageNodes) == 0 {
		fmt.Printf("No storage nodes found for %s, attempting third party fetch\n", objectKey)
		err = h.service.HandleThirdPartyFetchAndUpload(objectKey)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("Failed to fetch from third party: %v", err),
			})
			return
		}

		// 重新获取元数据
		metadata, err = h.service.GetMetadata(objectKey)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": fmt.Sprintf("Failed to get metadata after third party fetch: %v", err),
			})
			return
		}
	}

	// 从存储管理器读取文件
	fileObj, err := h.service.ReadFromStg1OrThirdParty(objectKey)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Failed to read object: %v", err),
		})
		return
	}

	// 设置响应头
	c.Header("Content-Type", fileObj.ContentType)
	c.Header("Content-Length", strconv.FormatInt(fileObj.Size, 10))
	c.Header("ETag", `"`+fileObj.MD5Hash+`"`)
	c.Header("Last-Modified", fileObj.CreatedAt.Format(http.TimeFormat))

	// 返回文件数据
	c.Data(http.StatusOK, fileObj.ContentType, fileObj.Data)
}

// GetObjectAPI 处理API GET对象请求
func (h *Handler) GetObjectAPI(c *gin.Context) {
	key := c.Param("key")

	metadata, err := h.service.GetMetadata(key)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Object not found"})
		return
	}

	fileObj, err := h.service.ReadFromStg1OrThirdParty(key)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"key":          fileObj.Key,
		"size":         fileObj.Size,
		"content_type": fileObj.ContentType,
		"md5_hash":     fileObj.MD5Hash,
		"created_at":   fileObj.CreatedAt,
		"metadata":     metadata,
	})
}