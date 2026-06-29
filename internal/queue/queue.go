package queue

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"lychee-go/internal/config"
	"lychee-go/internal/logger"

	"github.com/redis/go-redis/v9"
)

// ======== 任务定义 ========

// Job 任务接口（所有业务任务实现这个接口）
type Job interface {
	Handle() error // 任务执行逻辑
}

// JobWrapper 任务包装器（内部使用）
type JobWrapper struct {
	ID        string          `json:"id"`         // 任务 ID
	Name      string          `json:"name"`       // 任务名称
	Payload   json.RawMessage `json:"payload"`    // 任务参数（JSON）
	Attempts  int             `json:"attempts"`   // 已尝试次数
	MaxTries  int             `json:"max_tries"`  // 最大重试次数
	CreatedAt int64           `json:"created_at"` // 创建时间
}

// ======== 队列驱动接口 ========

type QueueDriver interface {
	Push(queue string, job *JobWrapper) error
	Pop(queue string) (*JobWrapper, error)
	Size(queue string) (int64, error)
	Clear(queue string) error
}

// ======== Redis 驱动实现 ========

type RedisDriver struct {
	client *redis.Client
	ctx    context.Context
	prefix string
}

func NewRedisDriver(client *redis.Client, prefix string) *RedisDriver {
	return &RedisDriver{
		client: client,
		ctx:    context.Background(),
		prefix: prefix,
	}
}

func (r *RedisDriver) key(queue string) string {
	return r.prefix + "queue:" + queue
}

func (r *RedisDriver) Push(queue string, job *JobWrapper) error {
	data, err := json.Marshal(job)
	if err != nil {
		return err
	}
	return r.client.LPush(r.ctx, r.key(queue), data).Err()
}

func (r *RedisDriver) Pop(queue string) (*JobWrapper, error) {
	// BRPop: 阻塞式弹出，超时 1 秒
	result, err := r.client.BRPop(r.ctx, time.Second, r.key(queue)).Result()
	if err == redis.Nil || len(result) == 0 {
		return nil, nil // 队列空
	}
	if err != nil {
		return nil, err
	}

	var job JobWrapper
	if err := json.Unmarshal([]byte(result[1]), &job); err != nil {
		return nil, err
	}
	return &job, nil
}

func (r *RedisDriver) Size(queue string) (int64, error) {
	return r.client.LLen(r.ctx, r.key(queue)).Result()
}

func (r *RedisDriver) Clear(queue string) error {
	return r.client.Del(r.ctx, r.key(queue)).Err()
}

// ======== 内存驱动（测试/开发用，无需 Redis） ========

type MemoryDriver struct {
	queues map[string][]*JobWrapper
	mu     sync.Mutex
}

func NewMemoryDriver() *MemoryDriver {
	return &MemoryDriver{
		queues: make(map[string][]*JobWrapper),
	}
}

func (m *MemoryDriver) Push(queue string, job *JobWrapper) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.queues[queue] = append([]*JobWrapper{job}, m.queues[queue]...)
	return nil
}

func (m *MemoryDriver) Pop(queue string) (*JobWrapper, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	jobs := m.queues[queue]
	if len(jobs) == 0 {
		time.Sleep(100 * time.Millisecond) // 模拟阻塞
		return nil, nil
	}

	job := jobs[len(jobs)-1]
	m.queues[queue] = jobs[:len(jobs)-1]
	return job, nil
}

func (m *MemoryDriver) Size(queue string) (int64, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	return int64(len(m.queues[queue])), nil
}

func (m *MemoryDriver) Clear(queue string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.queues, queue)
	return nil
}

// ======== 任务注册表（任务名称 → 构造函数） ========

type JobFactory func() Job

var (
	jobRegistry = make(map[string]JobFactory)
	registryMu  sync.RWMutex
)

// RegisterJob 注册任务类型（类似 ThinkPHP 的 Job 类）
// jobName: 任务名称，如 "send_email"
// factory: 返回一个空的 Job 实例（用于反序列化后调用 Handle）
func RegisterJob(jobName string, factory JobFactory) {
	registryMu.Lock()
	defer registryMu.Unlock()
	jobRegistry[jobName] = factory
	logger.Info("[Queue] Job registered: %s", jobName)
}

func getJobFactory(name string) (JobFactory, bool) {
	registryMu.RLock()
	defer registryMu.RUnlock()
	f, ok := jobRegistry[name]
	return f, ok
}

// ======== 队列管理器 ========

type Manager struct {
	driver QueueDriver
}

var manager *Manager

// Init 初始化队列系统
func Init(redisClient *redis.Client) error {
	driverType := config.GetString("queue.driver", "memory")
	prefix := config.GetString("queue.prefix", "lychee_go_")

	var driver QueueDriver
	switch driverType {
	case "redis":
		if redisClient == nil {
			return errors.New("redis client is nil for queue driver")
		}
		driver = NewRedisDriver(redisClient, prefix)
	default:
		driver = NewMemoryDriver()
	}

	manager = &Manager{driver: driver}
	logger.Info("[Queue] Initialized (driver: %s)", driverType)
	return nil
}

// ======== 对外 API ========

// Dispatch 投递任务到队列（异步执行）
// queueName: 队列名，如 "default", "emails"
// jobName: 任务名称（必须先用 RegisterJob 注册）
// payload: 任务参数（任意可 JSON 序列化的结构）
// maxTries: 最大重试次数（默认 3）
func Dispatch(queueName, jobName string, payload interface{}, maxTries ...int) error {
	if manager == nil {
		return errors.New("queue not initialized, call queue.Init() first")
	}

	tries := 3
	if len(maxTries) > 0 && maxTries[0] > 0 {
		tries = maxTries[0]
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("invalid payload: %w", err)
	}

	job := &JobWrapper{
		ID:        fmt.Sprintf("job_%d_%d", time.Now().UnixNano(), time.Now().Unix()),
		Name:      jobName,
		Payload:   payloadBytes,
		Attempts:  0,
		MaxTries:  tries,
		CreatedAt: time.Now().Unix(),
	}

	if err := manager.driver.Push(queueName, job); err != nil {
		return err
	}

	logger.Debug("[Queue] Dispatched: queue=%s, job=%s, id=%s", queueName, jobName, job.ID)
	return nil
}

// Size 查看队列长度
func Size(queueName string) (int64, error) {
	if manager == nil {
		return 0, errors.New("queue not initialized")
	}
	return manager.driver.Size(queueName)
}

// Clear 清空队列
func Clear(queueName string) error {
	if manager == nil {
		return errors.New("queue not initialized")
	}
	return manager.driver.Clear(queueName)
}

// ======== Worker（消费者） ========

type Worker struct {
	queueName string
	stopChan  chan struct{}
	running   bool
}

// NewWorker 创建一个消费者
func NewWorker(queueName string) *Worker {
	return &Worker{
		queueName: queueName,
		stopChan:  make(chan struct{}),
	}
}

// Start 启动消费者（阻塞执行，用 goroutine 调用）
func (w *Worker) Start() {
	if w.running {
		return
	}
	w.running = true
	logger.Info("[Queue] Worker started for queue: %s", w.queueName)

	for {
		// 检查是否停止
		select {
		case <-w.stopChan:
			logger.Info("[Queue] Worker stopped for queue: %s", w.queueName)
			return
		default:
		}

		// 弹出任务
		job, err := manager.driver.Pop(w.queueName)
		if err != nil {
			logger.Error("[Queue] Pop error: %v", err)
			time.Sleep(time.Second)
			continue
		}
		if job == nil {
			continue // 队列空，继续循环
		}

		// 执行任务
		w.processJob(job)
	}
}

// Stop 停止消费者
func (w *Worker) Stop() {
	if !w.running {
		return
	}
	close(w.stopChan)
	w.running = false
}

func (w *Worker) processJob(job *JobWrapper) {
	factory, ok := getJobFactory(job.Name)
	if !ok {
		logger.Error("[Queue] Unknown job type: %s", job.Name)
		return
	}

	jobInstance := factory()
	job.Attempts++

	logger.Info("[Queue] Processing job: %s (attempt %d/%d)", job.Name, job.Attempts, job.MaxTries)

	// 将 payload 反序列化到 Job 实例（约定：Job 结构的字段必须匹配 payload）
	if len(job.Payload) > 0 {
		if err := json.Unmarshal(job.Payload, jobInstance); err != nil {
			logger.Error("[Queue] Unmarshal payload failed: %v", err)
		}
	}

	// 执行任务
	if err := jobInstance.Handle(); err != nil {
		logger.Error("[Queue] Job failed: %s, error: %v", job.Name, err)

		// 如果还有重试次数，重新入队
		if job.Attempts < job.MaxTries {
			logger.Info("[Queue] Re-queue job for retry: %s", job.Name)
			_ = manager.driver.Push(w.queueName, job)
		}
		return
	}

	logger.Info("[Queue] Job completed successfully: %s", job.Name)
}

// StartWorker 便捷函数：启动一个后台消费者
func StartWorker(queueName string) *Worker {
	w := NewWorker(queueName)
	go w.Start()
	return w
}
