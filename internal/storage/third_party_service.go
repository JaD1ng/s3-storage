package storage

import (
	"fmt"
	"time"

	"mock-storage/internal/types"

	"github.com/google/uuid"
)

// MockThirdPartyService 模拟第三方服务实现
type MockThirdPartyService struct {
	name     string
	endpoint string
}

// NewMockThirdPartyService 创建模拟第三方服务
func NewMockThirdPartyService(name, endpoint string) *MockThirdPartyService {
	return &MockThirdPartyService{
		name:     name,
		endpoint: endpoint,
	}
}

// GetObject 从第三方服务获取对象
func (mts *MockThirdPartyService) GetObject(key string) (*types.FileObject, error) {
	fmt.Printf("[THIRD_PARTY] Fetching object from %s: %s\n", mts.endpoint, key)

	// 模拟网络延迟
	time.Sleep(200 * time.Millisecond)

	// 模拟从第三方服务获取数据
	// 在实际实现中，这里会发送HTTP请求到真实的第三方服务
	mockData := []byte(fmt.Sprintf("Mock data from third party service for key: %s", key))

	fileObj := &types.FileObject{
		ID:          uuid.New().String(),
		Key:         key,
		Size:        int64(len(mockData)),
		ContentType: "application/octet-stream",
		Data:        mockData,
		MD5Hash:     calculateMD5(mockData),
		CreatedAt:   time.Now(),
	}

	fmt.Printf("[THIRD_PARTY] Successfully fetched object: %s (size: %d bytes)\n", key, len(mockData))
	return fileObj, nil
}

// GetName 获取服务名称
func (mts *MockThirdPartyService) GetName() string {
	return mts.name
}
