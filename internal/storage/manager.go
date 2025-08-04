package storage

import (
	"fmt"
	"mock-storage/internal/types"
)

// ThirdPartyService 第三方服务接口
type ThirdPartyService interface {
	GetObject(key string) (*types.FileObject, error)
}

// Manager 存储管理器，管理多个存储节点
type Manager struct {
	nodes           []types.StorageNode
	thirdPartyService ThirdPartyService
}

// NewManager 创建存储管理器
func NewManager() *Manager {
	return &Manager{
		nodes: make([]types.StorageNode, 0),
	}
}

// AddNode 添加存储节点
func (sm *Manager) AddNode(node types.StorageNode) {
	sm.nodes = append(sm.nodes, node)
}

// SetThirdPartyService 设置第三方服务
func (sm *Manager) SetThirdPartyService(service ThirdPartyService) {
	sm.thirdPartyService = service
}

// WriteToAllNodes 顺序写入所有存储节点
func (sm *Manager) WriteToAllNodes(obj *types.FileObject) error {
	var lastErr error
	successCount := 0

	// 顺序写入每个节点
	for i, node := range sm.nodes {
		err := node.Write(obj)
		if err != nil {
			lastErr = err
			fmt.Printf("Failed to write to node %s: %v\n", node.GetNodeID(), err)
			continue
		}
		successCount++
		fmt.Printf("Step %d: Successfully wrote to node %s\n", i+1, node.GetNodeID())
	}

	// 如果至少有一个节点写入成功，则认为写入成功
	if successCount == 0 {
		return fmt.Errorf("failed to write to any storage node, last error: %v", lastErr)
	}

	if successCount < len(sm.nodes) {
		fmt.Printf("Warning: Only %d out of %d nodes wrote successfully\n", successCount, len(sm.nodes))
	}

	return nil
}

// ReadFromStg1OrThirdParty 优先从stg1读取，如果失败则从第三方获取
func (sm *Manager) ReadFromStg1OrThirdParty(key string) (*types.FileObject, error) {
	// 首先尝试从stg1读取
	if len(sm.nodes) > 0 {
		stg1Node := sm.nodes[0] // 假设第一个节点是stg1
		if stg1Node.GetNodeID() == "stg1" {
			obj, err := stg1Node.Read(key)
			if err == nil {
				fmt.Printf("Successfully read from stg1: %s\n", key)
				return obj, nil
			}
			fmt.Printf("Failed to read from stg1: %v\n", err)
		}
	}

	// 如果stg1读取失败，尝试从第三方服务获取
	if sm.thirdPartyService != nil {
		fmt.Printf("Attempting to fetch from third party service: %s\n", key)
		obj, err := sm.thirdPartyService.GetObject(key)
		if err != nil {
			return nil, fmt.Errorf("failed to get object from third party service: %v", err)
		}
		
		fmt.Printf("Successfully fetched from third party service: %s\n", key)
		return obj, nil
	}

	return nil, fmt.Errorf("failed to read file %s from stg1 and no third party service configured", key)
}

// ReadFromAnyNode 从任意一个可用的节点读取
func (sm *Manager) ReadFromAnyNode(key string) (*types.FileObject, error) {
	for _, node := range sm.nodes {
		obj, err := node.Read(key)
		if err == nil {
			return obj, nil
		}
		fmt.Printf("Failed to read from node %s: %v\n", node.GetNodeID(), err)
	}

	return nil, fmt.Errorf("failed to read file %s from any storage node", key)
}

// GetNodes 获取所有存储节点
func (sm *Manager) GetNodes() []types.StorageNode {
	return sm.nodes
}

// GetNodeIDs 获取所有节点ID
func (sm *Manager) GetNodeIDs() []string {
	ids := make([]string, len(sm.nodes))
	for i, node := range sm.nodes {
		ids[i] = node.GetNodeID()
	}
	return ids
}
