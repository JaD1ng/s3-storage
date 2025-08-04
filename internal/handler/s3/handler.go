package s3

import (
	"github.com/gin-gonic/gin"
)

// Handler S3接口处理器
type Handler struct {
	service *Service
}

// NewHandler 创建S3处理器
func NewHandler(service *Service) *Handler {
	return &Handler{
		service: service,
	}
}

// SetupRoutes 设置路由
func (h *Handler) SetupRoutes(router *gin.Engine) {
	// S3兼容的路由
	router.PUT("/:bucket/:key", h.PutObject)
	router.GET("/:bucket/:key", h.GetObject)
	router.DELETE("/:bucket/:key", h.DeleteObject)
	router.HEAD("/:bucket/:key", h.HeadObject)
	router.GET("/:bucket", h.ListObjects)

	// 管理接口
	api := router.Group("/api/v1")
	{
		api.GET("/objects", h.ListObjectsAPI)
		api.GET("/objects/:key", h.GetObjectAPI)
		api.POST("/objects", h.PutObjectAPI)
		api.DELETE("/objects/:key", h.DeleteObjectAPI)
		api.GET("/stats", h.GetStatsAPI)
		api.GET("/search", h.SearchObjectsAPI)
	}
}

// buildObjectKey 构建对象key（包含bucket前缀）
func (h *Handler) buildObjectKey(bucket, key string) string {
	return bucket + "/" + key
}

// extractObjectKey 从完整key中提取对象key（移除bucket前缀）
func (h *Handler) extractObjectKey(fullKey, bucketPrefix string) string {
	if len(fullKey) > len(bucketPrefix) {
		return fullKey[len(bucketPrefix):]
	}
	return fullKey
}