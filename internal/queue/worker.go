package queue

import (
	"fmt"
	"sync"
	"time"

	"mock-storage/internal/types"
)

// StorageManager 存储管理器接口（避免循环依赖）
type StorageManager interface {
	GetNodes() []types.StorageNode
}

// Worker 工作节点
type Worker struct {
	ID             string
	running        bool
	mutex          sync.RWMutex
	tasksProcessed int64
	processor      types.TaskProcessor
	storageManager StorageManager
}

// NewWorker 创建工作节点
func NewWorker(id string, processor types.TaskProcessor) *Worker {
	return &Worker{
		ID:        id,
		running:   false,
		processor: processor,
	}
}

// SetStorageManager 设置存储管理器
func (w *Worker) SetStorageManager(sm StorageManager) {
	w.storageManager = sm
}

// Start 启动工作节点
func (w *Worker) Start() {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.running = true
	fmt.Printf("[WORKER] Worker %s started\n", w.ID)
}

// Stop 停止工作节点
func (w *Worker) Stop() {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	w.running = false
	fmt.Printf("[WORKER] Worker %s stopped\n", w.ID)
}

// IsRunning 检查工作节点是否运行中
func (w *Worker) IsRunning() bool {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	return w.running
}

// ProcessTask 处理任务
func (w *Worker) ProcessTask(task *types.TaskMessage) error {
	w.mutex.Lock()
	w.tasksProcessed++
	w.mutex.Unlock()

	if w.processor != nil {
		return w.processor.Process(task)
	}

	// 默认处理逻辑
	return w.defaultProcess(task)
}

// GetTasksProcessed 获取已处理的任务数量
func (w *Worker) GetTasksProcessed() int64 {
	w.mutex.RLock()
	defer w.mutex.RUnlock()

	return w.tasksProcessed
}

// defaultProcess 默认任务处理逻辑
func (w *Worker) defaultProcess(task *types.TaskMessage) error {
	switch task.Type {
	case "upload_completed":
		return w.processUploadCompleted(task)
	case "cleanup":
		return w.processCleanup(task)
	case "replication_check":
		return w.processReplicationCheck(task)
	case "delete_from_storage":
		return w.processDeleteFromStorage(task)
	default:
		fmt.Printf("[WORKER] Unknown task type: %s\n", task.Type)
		return nil
	}
}

// processUploadCompleted 处理上传完成任务
func (w *Worker) processUploadCompleted(task *types.TaskMessage) error {
	fmt.Printf("[WORKER] Processing upload completed for object: %s\n", task.ObjectID)

	// 这里可以添加上传后的处理逻辑，比如：
	// - 发送通知
	// - 触发第三方服务
	// - 生成缩略图
	// - 病毒扫描等

	// 模拟处理时间
	time.Sleep(100 * time.Millisecond)

	return nil
}

// processCleanup 处理清理任务
func (w *Worker) processCleanup(task *types.TaskMessage) error {
	fmt.Printf("[WORKER] Processing cleanup for object: %s\n", task.ObjectID)

	// 这里可以添加清理逻辑，比如：
	// - 清理临时文件
	// - 清理过期数据
	// - 垃圾回收等

	time.Sleep(50 * time.Millisecond)

	return nil
}

// processReplicationCheck 处理副本检查任务
func (w *Worker) processReplicationCheck(task *types.TaskMessage) error {
	fmt.Printf("[WORKER] Processing replication check for object: %s\n", task.ObjectID)

	// 这里可以添加副本检查逻辑，比如：
	// - 检查数据完整性
	// - 验证副本一致性
	// - 修复损坏的副本等

	time.Sleep(200 * time.Millisecond)

	return nil
}

// processDeleteFromStorage 处理从存储节点删除任务
func (w *Worker) processDeleteFromStorage(task *types.TaskMessage) error {
	fmt.Printf("[WORKER] Processing delete from storage for object: %s\n", task.ObjectID)

	// 从任务数据中获取key
	key, ok := task.Data["key"].(string)
	if !ok {
		return fmt.Errorf("invalid key in delete task data")
	}

	// 检查是否有存储管理器
	if w.storageManager == nil {
		fmt.Printf("[WORKER] Warning: No storage manager available, cannot delete file %s\n", key)
		return fmt.Errorf("storage manager not available")
	}

	// 从所有存储节点删除文件
	fmt.Printf("[WORKER] Deleting file %s from all storage nodes\n", key)

	var lastError error
	deletedCount := 0

	for _, node := range w.storageManager.GetNodes() {
		err := node.Delete(key)
		if err != nil {
			fmt.Printf("[WORKER] Warning: failed to delete from node %s: %v\n", node.GetNodeID(), err)
			lastError = err
		} else {
			fmt.Printf("[WORKER] Successfully deleted from node %s: %s\n", node.GetNodeID(), key)
			deletedCount++
		}
	}

	// 如果至少有一个节点删除成功，认为任务成功
	if deletedCount > 0 {
		fmt.Printf("[WORKER] Delete task completed: %d nodes processed, %d successful\n", len(w.storageManager.GetNodes()), deletedCount)
		return nil
	}

	// 如果所有节点都删除失败，返回最后一个错误
	if lastError != nil {
		return fmt.Errorf("failed to delete from all storage nodes: %v", lastError)
	}

	return nil
}
