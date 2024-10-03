package stats

import (
	"fmt"
	"sort"
	"time"

	"github.com/tyxben/goloadtest/internal/worker"
)

type Stats struct {
	TotalRequests   int
	SuccessRequests int
	FailedRequests  int
	TotalDuration   time.Duration
	MinDuration     time.Duration
	MaxDuration     time.Duration
	AvgDuration     time.Duration
	Percentiles     map[float64]time.Duration
	StatusCodes     map[int]int
	ErrorTypes      map[string]int
	RequestsPerSec  float64
}

func NewStats() *Stats {
	return &Stats{
		MinDuration: time.Duration(1<<63 - 1),
		Percentiles: make(map[float64]time.Duration),
		StatusCodes: make(map[int]int),
		ErrorTypes:  make(map[string]int),
	}
}

func (s *Stats) AddResult(result worker.Result) {
	s.TotalRequests++
	if result.Error != nil {
		s.FailedRequests++
		errorType := fmt.Sprintf("%T", result.Error)
		s.ErrorTypes[errorType]++
	} else {
		s.SuccessRequests++
		s.TotalDuration += result.Duration
		s.StatusCodes[result.StatusCode]++

		if result.Duration < s.MinDuration {
			s.MinDuration = result.Duration
		}
		if result.Duration > s.MaxDuration {
			s.MaxDuration = result.Duration
		}
	}
}

func (s *Stats) CalculateStats(duration time.Duration) {
	if s.SuccessRequests > 0 {
		s.AvgDuration = s.TotalDuration / time.Duration(s.SuccessRequests)
	}
	s.RequestsPerSec = float64(s.TotalRequests) / duration.Seconds()

	// 计算百分位数
	var durations []time.Duration
	for code, count := range s.StatusCodes {
		for i := 0; i < count; i++ {
			durations = append(durations, time.Duration(code))
		}
	}
	sort.Slice(durations, func(i, j int) bool { return durations[i] < durations[j] })
	percentiles := []float64{50, 75, 90, 95, 99}
	for _, p := range percentiles {
		index := int(float64(len(durations)) * p / 100)
		if index > 0 && index < len(durations) {
			s.Percentiles[p] = durations[index]
		}
	}
}

func (s *Stats) Print() {
	fmt.Printf("测试完成:\n")
	fmt.Printf("总请求数: %d\n", s.TotalRequests)
	fmt.Printf("成功请求: %d\n", s.SuccessRequests)
	fmt.Printf("失败请求: %d\n", s.FailedRequests)
	fmt.Printf("请求成功率: %.2f%%\n", float64(s.SuccessRequests)/float64(s.TotalRequests)*100)
	fmt.Printf("总耗时: %v\n", s.TotalDuration)
	fmt.Printf("每秒请求数: %.2f\n", s.RequestsPerSec)
	fmt.Printf("最小响应时间: %v\n", s.MinDuration)
	fmt.Printf("最大响应时间: %v\n", s.MaxDuration)
	fmt.Printf("平均响应时间: %v\n", s.AvgDuration)

	fmt.Printf("\n响应时间分布:\n")
	for p, d := range s.Percentiles {
		fmt.Printf("%v%%分位数: %v\n", p, d)
	}

	fmt.Printf("\n状态码分布:\n")
	for code, count := range s.StatusCodes {
		fmt.Printf("状态码 %d: %d次\n", code, count)
	}

	fmt.Printf("\n错误类型分布:\n")
	for errType, count := range s.ErrorTypes {
		fmt.Printf("%s: %d次\n", errType, count)
	}
}
