package s3

import (
	"fmt"
	"time"

	"mock-storage/internal/metadata"
	"mock-storage/internal/queue"
	"mock-storage/internal/storage"
	"mock-storage/internal/types"
)

// Service S3业务逻辑服务
type Service struct {
	storageManager  *storage.Manager
	metadataService *metadata.MetaService
	queueManager    *queue.Manager
}

// NewService 创建S3业务服务
func NewService(storageManager *storage.Manager, metadataService *metadata.MetaService, queueManager *queue.Manager) *Service {
	return &Service{
		storageManager:  storageManager,
		metadataService: metadataService,
		queueManager:    queueManager,
	}
}

// ExecuteUploadFlow 执行完整的上传流程
func (s *Service) ExecuteUploadFlow(fileObj *types.FileObject) error {
	fmt.Printf("Starting upload flow for key: %s\n", fileObj.Key)

	// 步骤1-3: 顺序写入存储节点
	err := s.storageManager.WriteToAllNodes(fileObj)
	if err != nil {
		return fmt.Errorf("failed to write to storage nodes: %v", err)
	}

	// 步骤4: 写入元数据服务
	storageNodeIDs := s.storageManager.GetNodeIDs()
	err = s.metadataService.SaveMetadata(fileObj, storageNodeIDs)
	if err != nil {
		return fmt.Errorf("failed to save metadata: %v", err)
	}

	// 步骤5: 数据已经通过元数据服务保存到数据库

	// // 异步任务：发送到队列进行后续处理
	// task := &types.TaskMessage{
	// 	Type:     "upload_completed",
	// 	ObjectID: fileObj.ID,
	// 	Data: map[string]any{
	// 		"key":          fileObj.Key,
	// 		"size":         fileObj.Size,
	// 		"content_type": fileObj.ContentType,
	// 	},
	// 	CreatedAt: time.Now(),
	// }

	// err = s.queueManager.Enqueue(task)
	// if err != nil {
	// 	fmt.Printf("Warning: failed to enqueue task: %v\n", err)
	// 	// 不返回错误，因为主要流程已完成
	// }

	fmt.Printf("Upload flow completed successfully for key: %s\n", fileObj.Key)
	return nil
}

// HandleThirdPartyFetchAndUpload 处理从第三方获取并上传的逻辑
func (s *Service) HandleThirdPartyFetchAndUpload(objectKey string) error {
	// 从第三方服务获取对象
	fileObj, err := s.storageManager.ReadFromStg1OrThirdParty(objectKey)
	if err != nil {
		return fmt.Errorf("failed to fetch from third party: %v", err)
	}

	// 执行上传流程
	err = s.ExecuteUploadFlow(fileObj)
	if err != nil {
		return fmt.Errorf("failed to execute upload flow after third party fetch: %v", err)
	}

	return nil
}

// GetStg1Node 获取stg1存储节点
func (s *Service) GetStg1Node() types.StorageNode {
	nodes := s.storageManager.GetNodes()
	for _, node := range nodes {
		if node.GetNodeID() == "stg1" {
			return node
		}
	}
	// 如果没有找到stg1，返回第一个节点
	if len(nodes) > 0 {
		return nodes[0]
	}
	return nil
}

// DeleteFromAllStorageNodes 从所有存储节点删除文件
func (s *Service) DeleteFromAllStorageNodes(objectKey string) {
	for _, node := range s.storageManager.GetNodes() {
		err := node.Delete(objectKey)
		if err != nil {
			fmt.Printf("Warning: failed to delete from node %s: %v\n", node.GetNodeID(), err)
		} else {
			fmt.Printf("Successfully deleted from node %s: %s\n", node.GetNodeID(), objectKey)
		}
	}
}

// GetMetadata 获取对象元数据
func (s *Service) GetMetadata(objectKey string) (*types.MetadataEntry, error) {
	return s.metadataService.GetMetadata(objectKey)
}

// DeleteMetadata 删除对象元数据
func (s *Service) DeleteMetadata(objectKey string) error {
	return s.metadataService.DeleteMetadata(objectKey)
}

// ListMetadata 列出对象元数据
func (s *Service) ListMetadata(limit, offset int) ([]*types.MetadataEntry, error) {
	return s.metadataService.ListMetadata(limit, offset)
}

// ReadFromStg1OrThirdParty 从stg1或第三方读取对象
func (s *Service) ReadFromStg1OrThirdParty(objectKey string) (*types.FileObject, error) {
	return s.storageManager.ReadFromStg1OrThirdParty(objectKey)
}

// EnqueueDeleteTask 将删除任务加入队列
func (s *Service) EnqueueDeleteTask(objectKey string) error {
	task := &types.TaskMessage{
		Type:     "delete_from_storage",
		ObjectID: objectKey,
		Data: map[string]any{
			"key": objectKey,
		},
		CreatedAt: time.Now(),
	}

	return s.queueManager.Enqueue(task)
}

// GetStats 获取统计信息
func (s *Service) GetStats() (map[string]any, error) {
	return s.metadataService.GetStats()
}

// SearchMetadata 搜索元数据
func (s *Service) SearchMetadata(query string, limit int) ([]*types.MetadataEntry, error) {
	return s.metadataService.SearchMetadata(query, limit)
}
