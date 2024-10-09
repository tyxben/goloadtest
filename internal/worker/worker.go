package worker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/tyxben/goloadtest/pkg/config"
)

type Result struct {
	APIName    string
	StatusCode int
	Duration   time.Duration
	Error      error
	Response   json.RawMessage
}

// TestDataQueue 是一个线程安全的队列，用于存储测试数据
type TestDataQueue struct {
	data  []map[string]string
	mutex sync.Mutex
}

// NewTestDataQueue 创建一个新的 TestDataQueue 并预加载所有数据
func NewTestDataQueue(testData []map[string]string) *TestDataQueue {
	return &TestDataQueue{
		data: testData,
	}
}

// Next 返回队列中的下一个测试数据，如果队列为空则返回 nil
func (q *TestDataQueue) Next() map[string]string {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	if len(q.data) == 0 {
		return nil
	}

	item := q.data[0]
	q.data = q.data[1:]
	return item
}

func Run(cfg *config.Config, tasks <-chan struct{}, results chan<- Result, testDataQueue *TestDataQueue) {
	client := &http.Client{
		Timeout: time.Second * 10,
	}

	for range tasks {
		sessionData := make(map[string]interface{})
		testData := testDataQueue.Next()
		if testData == nil {
			asyncLog("警告: 所有测试数据已用完")
			break
		}
		for key, value := range testData {
			sessionData[key] = value
		}

		for _, apiName := range cfg.Workflow {
			apiConfig := cfg.APIs[apiName]
			result := callAPI(client, cfg.BaseURL+apiConfig.URL, apiConfig, sessionData)
			result.APIName = apiName
			results <- result

			if result.Error != nil {
				break
			}

			handleResponse(result.Response, apiConfig.Response, sessionData)
		}
	}
}

// 检查 API 是否需要测试数据
func needsTestData(apiConfig config.APIConfig) bool {
	for _, value := range apiConfig.Params {
		if strings.Contains(value, "{{walletAddr}}") || strings.Contains(value, "{{amount}}") {
			return true
		}
	}
	return false
}

func callAPI(client *http.Client, apiUrl string, apiConfig config.APIConfig, sessionData map[string]interface{}) Result {
	start := time.Now()

	// 准备查询参数
	if len(apiConfig.QueryParams) > 0 {
		queryParams := make(url.Values)
		for key, value := range apiConfig.QueryParams {
			if strings.HasPrefix(value, "{{") && strings.HasSuffix(value, "}}") {
				paramName := strings.Trim(value, "{}")
				if paramValue, ok := sessionData[paramName]; ok {
					queryParams.Set(key, fmt.Sprintf("%v", paramValue))
				}
			} else {
				queryParams.Set(key, value)
			}

		}
		apiUrl += "?" + queryParams.Encode()
	}

	// 准备请求体
	var body []byte
	if len(apiConfig.Body) > 0 {
		bodyMap := make(map[string]interface{})
		for key, value := range apiConfig.Body {
			if strings.HasPrefix(value, "{{") && strings.HasSuffix(value, "}}") {
				paramName := strings.Trim(value, "{}")
				if paramValue, ok := sessionData[paramName]; ok {
					bodyMap[key] = paramValue
				}
			} else {
				bodyMap[key] = value
			}
		}
		body, _ = json.Marshal(bodyMap)
	}

	req, err := http.NewRequest(apiConfig.Method, apiUrl, bytes.NewReader(body))
	if err != nil {
		asyncLog("创建请求失败: %v", err)
		return Result{Error: err}
	}
	req.Header.Set("Content-Type", "application/json")

	// 设置请求头
	for k, v := range apiConfig.Headers {
		req.Header.Set(k, replaceSessionData(v, sessionData))
	}

	resp, err := client.Do(req)
	if err != nil {
		asyncLog("发送请求失败: %v", err)
		return Result{Error: err}
	}
	defer resp.Body.Close()

	responseBody, _ := ioutil.ReadAll(resp.Body)
	duration := time.Since(start)
	//check if responseBody 包含code 且非 0 输出
	var responseMap map[string]interface{}
	err = json.Unmarshal(responseBody, &responseMap)
	if err != nil {
		asyncLog("警告: 无法解析响应 JSON: %v", err)
		return Result{Error: err}
	}
	if code, ok := responseMap["code"]; ok {
		// 将 code 转换为整数进行比较
		codeInt, isInt := code.(float64)
		if isInt && int(codeInt) != 0 {
			if walletAddr, ok := sessionData["walletAddr"]; ok {
				asyncLog("响应包含错误码: %v, 地址: %v", int(codeInt), walletAddr)
			}
		}
	}
	asyncLog("地址%s,响应: %v", sessionData["walletAddr"], string(responseBody))
	return Result{
		StatusCode: resp.StatusCode,
		Duration:   duration,
		Response:   responseBody,
	}
}

func handleResponse(response json.RawMessage, responseConfig map[string]string, sessionData map[string]interface{}) {

	// 解析整个 JSON 响应
	var responseMap map[string]interface{}
	err := json.Unmarshal(response, &responseMap)
	if err != nil {
		asyncLog("警告: 无法解析响应 JSON: %v", err)
		return
	}

	// 处理 responseConfig 中指定的所有字段
	for key, fieldName := range responseConfig {
		value := findFieldRecursively(responseMap, fieldName)
		if value != nil {
			sessionData[key] = value
		} else {
			asyncLog("警告: %s 在响应中未找到字段 %s", response, fieldName)

		}
	}
}

// 递归查找指定字段名的值
func findFieldRecursively(data interface{}, fieldName string) interface{} {
	switch v := data.(type) {
	case map[string]interface{}:
		for key, value := range v {
			if key == fieldName {
				return value
			}
			if result := findFieldRecursively(value, fieldName); result != nil {
				return result
			}
		}
	case []interface{}:
		for _, item := range v {
			if result := findFieldRecursively(item, fieldName); result != nil {
				return result
			}
		}
	}
	return nil
}

func replaceSessionData(value string, sessionData map[string]interface{}) string {
	for k, v := range sessionData {
		placeholder := "{{" + k + "}}"
		if strings.Contains(value, placeholder) {
			value = strings.ReplaceAll(value, placeholder, v.(string))
		}
	}
	return value
}

var (
	logChan chan string
	logWg   sync.WaitGroup
)

func init() {
	logChan = make(chan string, 1000) // 缓冲区大小可以根据需要调整
	logWg.Add(1)
	go logWriter()
}

func logWriter() {
	defer logWg.Done()
	logger := log.New(os.Stderr, "", log.LstdFlags)
	for msg := range logChan {
		logger.Println(msg)
	}
}

func asyncLog(format string, v ...interface{}) {
	select {
	case logChan <- fmt.Sprintf(format, v...):
	default:
		// 如果通道已满，丢弃日志消息
	}
}
