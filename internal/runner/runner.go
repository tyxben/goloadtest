package runner

import (
	"log"
	"sync"
	"time"

	"github.com/tyxben/goloadtest/internal/stats"
	"github.com/tyxben/goloadtest/internal/worker"
	"github.com/tyxben/goloadtest/pkg/config"
)

type Runner struct {
	Config *config.Config
	Stats  *stats.Stats
}

func NewRunner(cfg *config.Config) *Runner {
	return &Runner{
		Config: cfg,
		Stats:  stats.NewStats(),
	}
}

func (r *Runner) Run() {
	log.Println("开始运行测试...")
	tasks := make(chan struct{}, r.Config.Concurrency)
	results := make(chan worker.Result)

	var wg sync.WaitGroup
	for i := 0; i < r.Config.Concurrency; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			log.Printf("启动工作协程 #%d", index)
			worker.Run(r.Config, tasks, results)
		}(i)
	}

	// 启动任务生成器
	go r.generateTasks(tasks)

	// 启动结果收集器
	go func() {
		wg.Wait()
		close(results)
		log.Println("所有工作协程完成，关闭结果通道")
	}()

	// 收集结果
	log.Println("开始收集结果...")
	startTime := time.Now()
	for result := range results {
		r.Stats.AddResult(result)
	}
	duration := time.Since(startTime)
	log.Printf("测试完成，总耗时: %v", duration)

	// 计算最终统计信息
	r.Stats.CalculateStats(duration)
}

func (r *Runner) generateTasks(tasks chan<- struct{}) {
	log.Println("开始生成任务...")
	startTime := time.Now()

	if r.Config.TotalRequests > 0 {
		// 按照指定次数生成任务
		for i := 0; i < r.Config.TotalRequests; i++ {
			tasks <- struct{}{}
		}
	} else {
		// 按照持续时间生成任务
		for time.Since(startTime) < time.Duration(r.Config.Duration)*time.Second {
			select {
			case tasks <- struct{}{}:
			default:
				// Channel 已满，跳过本次发送
			}
		}
	}

	close(tasks)
	log.Println("任务生成完成，关闭任务通道")
}
