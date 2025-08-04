package queue

import (
	"fmt"
	"sync"

	"mock-storage/internal/types"
)

// Manager 队列管理器
type Manager struct {
	queue     chan *types.TaskMessage
	workers   []*Worker
	maxSize   int
	mutex     sync.RWMutex
	running   bool
	waitGroup sync.WaitGroup
}

// NewManager 创建队列管理器
func NewManager(maxSize int) *Manager {
	return &Manager{
		queue:   make(chan *types.TaskMessage, maxSize),
		maxSize: maxSize,
		workers: make([]*Worker, 0),
		running: false,
	}
}

// Start 启动队列管理器
func (qm *Manager) Start() error {
	qm.mutex.Lock()
	defer qm.mutex.Unlock()

	if qm.running {
		return fmt.Errorf("queue manager is already running")
	}

	qm.running = true

	// 为所有已添加的Worker启动处理循环
	for _, worker := range qm.workers {
		qm.waitGroup.Add(1)
		go qm.runWorker(worker)
	}

	fmt.Printf("[QUEUE] Queue manager started with %d workers\n", len(qm.workers))
	return nil
}

// Stop 停止队列管理器
func (qm *Manager) Stop() error {
	qm.mutex.Lock()
	defer qm.mutex.Unlock()

	if !qm.running {
		return fmt.Errorf("queue manager is not running")
	}

	qm.running = false

	// 停止所有工作节点
	for _, worker := range qm.workers {
		worker.Stop()
	}

	// 关闭队列通道
	close(qm.queue)

	// 等待所有工作节点完成
	qm.waitGroup.Wait()

	fmt.Println("[QUEUE] Queue manager stopped")
	return nil
}

// Enqueue 将任务加入队列
func (qm *Manager) Enqueue(task *types.TaskMessage) error {
	qm.mutex.RLock()
	defer qm.mutex.RUnlock()

	if !qm.running {
		return fmt.Errorf("queue manager is not running")
	}

	select {
	case qm.queue <- task:
		fmt.Printf("[QUEUE] Task enqueued: %s (ID: %s)\n", task.Type, task.ObjectID)
		return nil
	default:
		return fmt.Errorf("queue is full")
	}
}

// AddWorker 添加工作节点
func (qm *Manager) AddWorker(worker *Worker) {
	qm.mutex.Lock()
	defer qm.mutex.Unlock()

	qm.workers = append(qm.workers, worker)

	if qm.running {
		qm.waitGroup.Add(1)
		go qm.runWorker(worker)
	}
}

// runWorker 运行单个工作节点
func (qm *Manager) runWorker(worker *Worker) {
	defer qm.waitGroup.Done()

	fmt.Printf("[QUEUE] Worker %s started\n", worker.ID)

	for {
		select {
		case task, ok := <-qm.queue:
			if !ok {
				// 队列已关闭
				fmt.Printf("[QUEUE] Worker %s stopped (queue closed)\n", worker.ID)
				return
			}

			if !worker.IsRunning() {
				// 工作节点已停止
				fmt.Printf("[QUEUE] Worker %s stopped\n", worker.ID)
				return
			}

			// 处理任务
			err := worker.ProcessTask(task)
			if err != nil {
				fmt.Printf("[QUEUE] Worker %s failed to process task %s: %v\n",
					worker.ID, task.ObjectID, err)
			} else {
				fmt.Printf("[QUEUE] Worker %s completed task %s\n",
					worker.ID, task.ObjectID)
			}
		}
	}
}

// GetStats 获取队列统计信息
func (qm *Manager) GetStats() map[string]any {
	qm.mutex.RLock()
	defer qm.mutex.RUnlock()

	stats := map[string]any{
		"running":       qm.running,
		"queue_size":    len(qm.queue),
		"max_size":      qm.maxSize,
		"worker_count":  len(qm.workers),
		"capacity_used": float64(len(qm.queue)) / float64(qm.maxSize) * 100,
	}

	// 工作节点状态
	workerStats := make([]map[string]any, len(qm.workers))
	for i, worker := range qm.workers {
		workerStats[i] = map[string]any{
			"id":              worker.ID,
			"running":         worker.IsRunning(),
			"tasks_processed": worker.GetTasksProcessed(),
		}
	}
	stats["workers"] = workerStats

	return stats
}
