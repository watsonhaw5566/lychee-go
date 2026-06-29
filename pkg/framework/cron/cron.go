package cron

import (
	"errors"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/watsonhaw5566/lychee-go/pkg/framework/logger"
)

// ======== 定时任务接口 ========

// Task 任务接口
type Task interface {
	Name() string // 任务名称
	Spec() string // Cron 表达式，如 "*/5 * * * * *"
	Run() error   // 执行任务
}

// TaskFunc 函数类型任务（便捷）
type TaskFunc struct {
	taskName string
	spec     string
	fn       func() error
}

func (tf *TaskFunc) Name() string { return tf.taskName }
func (tf *TaskFunc) Spec() string { return tf.spec }
func (tf *TaskFunc) Run() error   { return tf.fn() }

// NewFunc 创建一个函数式任务
func NewFunc(name, spec string, fn func() error) Task {
	return &TaskFunc{taskName: name, spec: spec, fn: fn}
}

// ======== Cron 表达式解析器 ========
// 支持 6 位格式: 秒 分 时 日 月 周
// 如:
//   "* * * * * *"        每秒执行
//   "*/5 * * * * *"      每 5 秒执行
//   "0 */10 * * * *"     每 10 分钟的第 0 秒执行
//   "0 0 2 * * *"        每天凌晨 2 点执行
//   "0 0 0 * * 1"        每周一凌晨执行

type cronField struct {
	min, max int
	value    []int
}

type cronSchedule struct {
	fields [6]cronField
	next   time.Time
}

func parseField(field string, min, max int) ([]int, error) {
	var result []int

	// 单个 *
	if field == "*" {
		for i := min; i <= max; i++ {
			result = append(result, i)
		}
		return result, nil
	}

	// 范围: 1-5
	if strings.Contains(field, "-") && !strings.Contains(field, "/") {
		parts := strings.SplitN(field, "-", 2)
		start, _ := atoi(parts[0])
		end, _ := atoi(parts[1])
		for i := start; i <= end; i++ {
			if i >= min && i <= max {
				result = append(result, i)
			}
		}
		return result, nil
	}

	// 步进: */5 或 0-30/10
	if strings.Contains(field, "/") {
		parts := strings.SplitN(field, "/", 2)
		step, _ := atoi(parts[1])
		base := parts[0]

		start := min
		end := max
		if strings.Contains(base, "-") {
			rangeParts := strings.SplitN(base, "-", 2)
			start, _ = atoi(rangeParts[0])
			end, _ = atoi(rangeParts[1])
		} else if base != "*" {
			start, _ = atoi(base)
		}

		for i := start; i <= end; i += step {
			if i >= min && i <= max {
				result = append(result, i)
			}
		}
		return result, nil
	}

	// 列表: 1,3,5
	if strings.Contains(field, ",") {
		parts := strings.Split(field, ",")
		for _, p := range parts {
			v, _ := atoi(p)
			if v >= min && v <= max {
				result = append(result, v)
			}
		}
		return result, nil
	}

	// 单个数值
	v, err := atoi(field)
	if err != nil {
		return nil, err
	}
	if v >= min && v <= max {
		result = append(result, v)
	}
	return result, nil
}

// 简易 atoi（避免引入 strconv 到多个 import）
func atoi(s string) (int, error) {
	if s == "" {
		return 0, errors.New("empty")
	}
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, errors.New("invalid digit")
		}
		n = n*10 + int(c-'0')
	}
	return n, nil
}

// parseCron 解析 Cron 表达式
func parseCron(spec string) (*cronSchedule, error) {
	fields := strings.Fields(spec)
	if len(fields) != 6 {
		return nil, errors.New("cron expression must have 6 fields: second minute hour day month weekday")
	}

	ranges := [6][2]int{
		{0, 59}, // second
		{0, 59}, // minute
		{0, 23}, // hour
		{1, 31}, // day
		{1, 12}, // month
		{0, 6},  // weekday (0 = Sunday)
	}

	var sched cronSchedule
	for i, f := range fields {
		values, err := parseField(f, ranges[i][0], ranges[i][1])
		if err != nil {
			return nil, err
		}
		sched.fields[i] = cronField{min: ranges[i][0], max: ranges[i][1], value: values}
		sort.Ints(sched.fields[i].value)
	}

	return &sched, nil
}

// Next 计算下一次执行时间
func (s *cronSchedule) Next(from time.Time) time.Time {
	t := from.Truncate(time.Second).Add(time.Second)

	// 最多尝试 366 天
	for days := 0; days < 366; days++ {
		for hour := 0; hour < 24; hour++ {
			for minute := 0; minute < 60; minute++ {
				for second := 0; second < 60; second++ {
					candidate := time.Date(
						t.Year(), t.Month(), t.Day(),
						hour, minute, second, 0, t.Location(),
					)
					if candidate.Before(t) {
						continue
					}

					// 检查各字段是否匹配
					secMatch := contains(s.fields[0].value, second)
					minMatch := contains(s.fields[1].value, minute)
					hourMatch := contains(s.fields[2].value, hour)
					dayMatch := contains(s.fields[3].value, candidate.Day())
					monthMatch := contains(s.fields[4].value, int(candidate.Month()))
					weekdayMatch := contains(s.fields[5].value, int(candidate.Weekday()))

					if secMatch && minMatch && hourMatch && dayMatch && monthMatch && weekdayMatch {
						return candidate
					}
				}
			}
		}
		t = time.Date(t.Year(), t.Month(), t.Day()+1, 0, 0, 0, 0, t.Location())
	}
	return from.Add(24 * time.Hour)
}

func contains(slice []int, val int) bool {
	for _, v := range slice {
		if v == val {
			return true
		}
	}
	return false
}

// ======== Cron 调度器 ========

type Cron struct {
	tasks    []Task
	scheds   map[*Task]*cronSchedule
	stopChan chan struct{}
	mu       sync.Mutex
	running  bool
}

func New() *Cron {
	return &Cron{
		scheds:   make(map[*Task]*cronSchedule),
		stopChan: make(chan struct{}),
	}
}

// Add 添加任务
func (c *Cron) Add(task Task) error {
	sched, err := parseCron(task.Spec())
	if err != nil {
		return err
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.tasks = append(c.tasks, task)
	c.scheds[&task] = sched

	logger.Info("[Cron] Task added: %s (spec: %s)", task.Name(), task.Spec())
	return nil
}

// AddFunc 便捷方法：添加函数式任务
func (c *Cron) AddFunc(name, spec string, fn func() error) error {
	return c.Add(NewFunc(name, spec, fn))
}

// Start 启动调度器（异步）
func (c *Cron) Start() {
	c.mu.Lock()
	if c.running {
		c.mu.Unlock()
		return
	}
	c.running = true
	c.mu.Unlock()

	logger.Info("[Cron] Scheduler started (%d tasks)", len(c.tasks))

	go func() {
		ticker := time.NewTicker(time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-c.stopChan:
				logger.Info("[Cron] Scheduler stopped")
				return
			case now := <-ticker.C:
				c.checkAndRun(now)
			}
		}
	}()
}

// Stop 停止调度器
func (c *Cron) Stop() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if !c.running {
		return
	}
	close(c.stopChan)
	c.stopChan = make(chan struct{})
	c.running = false
}

// checkAndRun 每秒检查一次，执行到达时间的任务
func (c *Cron) checkAndRun(now time.Time) {
	c.mu.Lock()
	tasks := make([]Task, len(c.tasks))
	copy(tasks, c.tasks)
	c.mu.Unlock()

	for _, task := range tasks {
		sched := c.scheds[&task]
		if sched == nil {
			continue
		}

		// 检查当前秒是否匹配
		truncated := now.Truncate(time.Second)
		nextTime := sched.Next(truncated.Add(-time.Second))

		// 如果下一次执行时间就是当前这一秒（±1 秒容差）
		diff := nextTime.Sub(truncated)
		if diff >= -500*time.Millisecond && diff <= 500*time.Millisecond {
			go func(t Task) {
				defer func() {
					if r := recover(); r != nil {
						logger.Error("[Cron] Task panic: %s, recover: %v", t.Name(), r)
					}
				}()

				start := time.Now()
				logger.Info("[Cron] Running task: %s", t.Name())
				if err := t.Run(); err != nil {
					logger.Error("[Cron] Task failed: %s, error: %v", t.Name(), err)
				} else {
					logger.Info("[Cron] Task completed: %s (took %v)", t.Name(), time.Since(start))
				}
			}(task)
		}
	}
}

// ======== 全局实例 ========

var globalCron = New()

// Add 全局注册任务
func Add(task Task) error {
	return globalCron.Add(task)
}

// AddFunc 全局注册函数任务
func AddFunc(name, spec string, fn func() error) error {
	return globalCron.AddFunc(name, spec, fn)
}

// Start 启动全局调度器
func Start() {
	globalCron.Start()
}

// Stop 停止全局调度器
func Stop() {
	globalCron.Stop()
}

// TaskCount 获取任务总数
func TaskCount() int {
	globalCron.mu.Lock()
	defer globalCron.mu.Unlock()
	return len(globalCron.tasks)
}
