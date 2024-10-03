package worker

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/tidwall/gjson"
	"github.com/tyxben/goloadtest/pkg/config"
)

type Result struct {
	APIName    string
	StatusCode int
	Duration   time.Duration
	Error      error
	Response   json.RawMessage
}

func Run(cfg *config.Config, tasks <-chan struct{}, results chan<- Result) {
	client := &http.Client{
		Timeout: time.Second * 10,
	}

	testDataIndex := 0
	for range tasks {
		sessionData := make(map[string]interface{})
		if len(cfg.TestData) > 0 {
			if testDataIndex >= len(cfg.TestData) {
				testDataIndex = 0 // 循环使用测试数据
			}
			// 从测试数据中读取所有列
			for key, value := range cfg.TestData[testDataIndex] {
				sessionData[key] = value
			}
			testDataIndex++
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
	log.Printf("apiConfig.QueryParams: %v", apiConfig.QueryParams)
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
		log.Printf("创建请求失败: %v", err)
		return Result{Error: err}
	}

	// 设置请求头
	for k, v := range apiConfig.Headers {
		req.Header.Set(k, replaceSessionData(v, sessionData))
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("发送请求失败: %v", err)
		return Result{Error: err}
	}
	defer resp.Body.Close()

	responseBody, _ := ioutil.ReadAll(resp.Body)
	duration := time.Since(start)
	log.Printf("API 响应时间: %v", duration)

	return Result{
		StatusCode: resp.StatusCode,
		Duration:   duration,
		Response:   responseBody,
	}
}

func handleResponse(response json.RawMessage, responseConfig map[string]string, sessionData map[string]interface{}) {
	for key, jsonPath := range responseConfig {
		result := gjson.GetBytes(response, jsonPath)
		if !result.Exists() {
			log.Printf("警告: 在响应中未找到路径 %s", jsonPath)
			continue
		}

		value := result.String()
		if value != "" {
			sessionData[key] = value
		} else {
			log.Printf("警告: 提取的值为空")
		}
	}
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

func prepareBody(bodyTemplate json.RawMessage, sessionData map[string]interface{}) []byte {
	if len(bodyTemplate) == 0 {
		return nil
	}

	var bodyMap map[string]interface{}
	json.Unmarshal(bodyTemplate, &bodyMap)

	for k, v := range bodyMap {
		if strValue, ok := v.(string); ok && strValue == "{{sessionData}}" {
			bodyMap[k] = sessionData[k]
		}
	}

	body, _ := json.Marshal(bodyMap)
	return body
}
