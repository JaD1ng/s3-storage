package s3

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// DeleteObject 处理DELETE对象请求
func (h *Handler) DeleteObject(c *gin.Context) {
	bucket := c.Param("bucket")
	key := c.Param("key")

	// 构建对象key（包含bucket前缀）
	objectKey := h.buildObjectKey(bucket, key)

	// 从元数据服务删除
	err := h.service.DeleteMetadata(objectKey)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": fmt.Sprintf("Object not found: %v", err),
		})
		return
	}

	// 异步删除存储节点中的文件
	err = h.service.EnqueueDeleteTask(objectKey)
	if err != nil {
		fmt.Printf("Warning: failed to enqueue delete task: %v\n", err)
		// 不返回错误，因为元数据已删除
	}

	// 返回成功响应
	c.Status(http.StatusNoContent)
}

// DeleteObjectAPI 处理API DELETE对象请求
func (h *Handler) DeleteObjectAPI(c *gin.Context) {
	key := c.Param("key")

	err := h.service.DeleteMetadata(key)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Object not found"})
		return
	}

	// 异步删除存储节点中的文件
	err = h.service.EnqueueDeleteTask(key)
	if err != nil {
		fmt.Printf("Warning: failed to enqueue delete task: %v\n", err)
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Object deleted successfully",
	})
}