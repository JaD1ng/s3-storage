package service

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"mock-storage/internal/config"
	"mock-storage/internal/handler/s3"
	"mock-storage/internal/metadata"
	"mock-storage/internal/queue"
	"mock-storage/internal/storage"

	"github.com/gin-gonic/gin"
)

// ObjectStorageService 对象存储服务
type ObjectStorageService struct {
	config          *config.Config
	storageManager  *storage.Manager
	databaseManager *metadata.DatabaseManager
	metadataService *metadata.MetaService
	queueManager    *queue.Manager
	s3Handler       *s3.Handler
	server          *http.Server
}

// NewObjectStorageService 创建对象存储服务
func NewObjectStorageService() (*ObjectStorageService, error) {
	// 加载配置
	cfg, err := config.Load("config.json")
	if err != nil {
		return nil, fmt.Errorf("failed to load config: %v", err)
	}

	service := &ObjectStorageService{
		config: cfg,
	}

	// 初始化各个组件
	err = service.initializeComponents()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize components: %v", err)
	}

	return service, nil
}

// ensureDataDirectories 确保所有必要的数据目录存在
func (oss *ObjectStorageService) ensureDataDirectories() error {
	// 1. 确保主数据目录存在
	dataDir := oss.config.Storage.DataDir
	err := os.MkdirAll(dataDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create data directory %s: %v", dataDir, err)
	}
	fmt.Printf("- 确保数据目录存在: %s\n", dataDir)

	// 2. 确保数据库文件目录存在
	dbPath := oss.config.Database.DSN
	dbDir := filepath.Dir(dbPath)
	err = os.MkdirAll(dbDir, 0755)
	if err != nil {
		return fmt.Errorf("failed to create database directory %s: %v", dbDir, err)
	}
	fmt.Printf("- 确保数据库目录存在: %s\n", dbDir)

	// 3. 确保所有存储节点目录存在
	for _, nodeConfig := range oss.config.Storage.Nodes {
		err = os.MkdirAll(nodeConfig.Path, 0755)
		if err != nil {
			return fmt.Errorf("failed to create storage node directory %s: %v", nodeConfig.Path, err)
		}
		fmt.Printf("- 确保存储节点目录存在: %s (%s)\n", nodeConfig.ID, nodeConfig.Path)
	}

	return nil
}

// initializeComponents 初始化所有组件
func (oss *ObjectStorageService) initializeComponents() error {
	fmt.Println("=== 初始化对象存储服务 ===")

	// 0. 确保数据目录存在
	fmt.Println("检查并创建数据目录...")
	err := oss.ensureDataDirectories()
	if err != nil {
		return fmt.Errorf("failed to ensure data directories: %v", err)
	}

	// 1. 初始化数据库管理器
	fmt.Println("初始化数据库...")
	oss.databaseManager, err = metadata.NewDatabaseManager(oss.config.Database.Driver, oss.config.Database.DSN)
	if err != nil {
		return fmt.Errorf("failed to initialize database: %v", err)
	}

	// 2. 初始化存储管理器
	fmt.Println("初始化存储节点...")
	oss.storageManager = storage.NewManager()

	// 创建存储节点
	for _, nodeConfig := range oss.config.Storage.Nodes {
		node, err := storage.NewFileStorageNode(nodeConfig.ID, nodeConfig.Path)
		if err != nil {
			return fmt.Errorf("failed to create storage node %s: %v", nodeConfig.ID, err)
		}
		oss.storageManager.AddNode(node)
		fmt.Printf("- 创建存储节点: %s (%s)\n", nodeConfig.ID, nodeConfig.Path)
	}

	// 设置第三方服务
	fmt.Println("初始化第三方服务...")
	thirdPartyService := storage.NewMockThirdPartyService("mock-third-party", "http://mock-third-party.example.com/api")
	oss.storageManager.SetThirdPartyService(thirdPartyService)
	fmt.Printf("- 设置第三方服务: %s\n", thirdPartyService.GetName())

	// 3. 初始化元数据服务
	fmt.Println("初始化元数据服务...")
	oss.metadataService = metadata.NewMetaService(oss.databaseManager)

	// 4. 初始化队列管理器
	fmt.Println("初始化队列管理器...")
	oss.queueManager = queue.NewManager(oss.config.Queue.Size)

	// 创建工作节点
	worker1 := queue.NewWorker("worker-1", nil) // 使用默认处理器
	worker2 := queue.NewWorker("worker-2", nil) // 使用默认处理器

	// 为工作节点设置存储管理器，使其能够执行文件删除操作
	worker1.SetStorageManager(oss.storageManager)
	worker2.SetStorageManager(oss.storageManager)

	oss.queueManager.AddWorker(worker1)
	oss.queueManager.AddWorker(worker2)

	worker1.Start()
	worker2.Start()

	// 5. 初始化S3处理器
	fmt.Println("初始化S3接口处理器...")
	s3Service := s3.NewService(oss.storageManager, oss.metadataService, oss.queueManager)
	oss.s3Handler = s3.NewHandler(s3Service)

	fmt.Println("=== 所有组件初始化完成 ===")
	return nil
}

// Start 启动服务
func (oss *ObjectStorageService) Start() error {
	fmt.Printf("启动对象存储服务在 %s:%s\n", oss.config.Server.Host, oss.config.Server.Port)

	// 启动队列管理器
	err := oss.queueManager.Start()
	if err != nil {
		return fmt.Errorf("failed to start queue manager: %v", err)
	}

	// 设置Gin模式
	gin.SetMode(gin.ReleaseMode)

	// 创建路由器
	router := gin.New()
	router.Use(gin.Logger())
	router.Use(gin.Recovery())

	// 添加CORS中间件
	router.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, HEAD, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Origin, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// 设置路由
	oss.s3Handler.SetupRoutes(router)

	// 健康检查
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// 管理API
	api := router.Group("/api/v1")
	{
		// TODO: 实现管理API方法
		api.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})
	}

	// 创建HTTP服务器
	oss.server = &http.Server{
		Addr:    oss.config.Server.Host + ":" + oss.config.Server.Port,
		Handler: router,
	}

	// 启动服务器
	go func() {
		if err := oss.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			fmt.Printf("服务器启动失败: %v\n", err)
		}
	}()

	fmt.Printf("对象存储服务已启动: http://%s:%s\n", oss.config.Server.Host, oss.config.Server.Port)
	fmt.Println("\n可用的端点:")
	fmt.Println("  - PUT /{bucket}/{key}     - 上传对象")
	fmt.Println("  - GET /{bucket}/{key}     - 下载对象")
	fmt.Println("  - DELETE /{bucket}/{key}  - 删除对象")
	fmt.Println("  - HEAD /{bucket}/{key}    - 获取对象元数据")
	fmt.Println("  - GET /{bucket}           - 列出对象")
	fmt.Println("  - GET /health             - 健康检查")
	fmt.Println("  - GET /api/v1/objects     - 管理API")
	fmt.Println()

	return nil
}

// Stop 停止服务
func (oss *ObjectStorageService) Stop() error {
	fmt.Println("正在停止对象存储服务...")

	// 停止HTTP服务器
	if oss.server != nil {
		if err := oss.server.Close(); err != nil {
			fmt.Printf("停止HTTP服务器时出错: %v\n", err)
		}
	}

	// 停止队列管理器
	if oss.queueManager != nil {
		if err := oss.queueManager.Stop(); err != nil {
			fmt.Printf("停止队列管理器时出错: %v\n", err)
		}
	}

	// 关闭数据库
	if oss.databaseManager != nil {
		if err := oss.databaseManager.Close(); err != nil {
			fmt.Printf("关闭数据库时出错: %v\n", err)
		}
	}

	fmt.Println("✅ 对象存储服务已停止")
	return nil
}
