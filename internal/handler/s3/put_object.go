package s3

import (
	"fmt"
	"io"
	"net/http"
	"time"

	"mock-storage/internal/types"
	"mock-storage/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// PutObject 处理PUT对象请求
func (h *Handler) PutObject(c *gin.Context) {
	bucket := c.Param("bucket")
	key := c.Param("key")

	// 读取请求体
	data, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": fmt.Sprintf("Failed to read request body: %v", err),
		})
		return
	}

	// 获取Content-Type
	contentType := c.GetHeader("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// 构建对象key（包含bucket前缀）
	objectKey := h.buildObjectKey(bucket, key)

	// 创建文件对象
	fileObj := &types.FileObject{
		ID:          uuid.New().String(),
		Key:         objectKey,
		Size:        int64(len(data)),
		ContentType: contentType,
		Data:        data,
		MD5Hash:     utils.CalculateMD5(data),
		CreatedAt:   time.Now(),
	}

	// 执行完整的上传流程
	err = h.service.ExecuteUploadFlow(fileObj)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": fmt.Sprintf("Upload failed: %v", err),
		})
		return
	}

	// 返回成功响应
	c.Header("ETag", `"`+fileObj.MD5Hash+`"`)
	c.JSON(http.StatusOK, types.UploadResponse{
		Success:  true,
		ObjectID: fileObj.ID,
		Key:      objectKey,
		Size:     fileObj.Size,
		MD5Hash:  fileObj.MD5Hash,
		Message:  "Object uploaded successfully",
	})
}

// PutObjectAPI 处理API PUT对象请求
func (h *Handler) PutObjectAPI(c *gin.Context) {
	var req types.UploadRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fileObj := &types.FileObject{
		ID:          uuid.New().String(),
		Key:         req.Key,
		Size:        int64(len(req.Data)),
		ContentType: req.ContentType,
		Data:        req.Data,
		MD5Hash:     utils.CalculateMD5(req.Data),
		CreatedAt:   time.Now(),
	}

	err := h.service.ExecuteUploadFlow(fileObj)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, types.UploadResponse{
		Success:  true,
		ObjectID: fileObj.ID,
		Key:      fileObj.Key,
		Size:     fileObj.Size,
		MD5Hash:  fileObj.MD5Hash,
		Message:  "Object uploaded successfully",
	})
}