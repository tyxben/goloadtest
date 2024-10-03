package config

import (
	"encoding/csv"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
)

type APIConfig struct {
	URL         string            `json:"url"`
	Method      string            `json:"method"`
	Headers     map[string]string `json:"headers"`
	Body        map[string]string `json:"body"`
	QueryParams map[string]string `json:"queryParams"`
	Response    map[string]string `json:"response"`
	Params      []string          `json:"params"`
}

type Config struct {
	TotalRequests int                  `json:"totalRequests"`
	Concurrency   int                  `json:"concurrency"`
	Duration      int                  `json:"duration"`
	Workflow      []string             `json:"workflow"`
	TokenHeader   string               `json:"tokenHeader"`
	BaseURL       string               `json:"baseURL"`
	APIs          map[string]APIConfig `json:"apis"`
	TestData      []map[string]string
}

func Parse() (*Config, error) {
	configFile := flag.String("config", "config.json", "配置文件路径")
	apiFile := flag.String("api", "api.json", "API配置文件路径")
	testDataFile := flag.String("testdata", "", "测试数据 CSV 文件路径（可选）")
	flag.Parse()

	cfg, err := loadFromFile(*configFile)
	if err != nil {
		return nil, fmt.Errorf("加载配置文件失败: %w", err)
	}

	apis, err := loadAPIsFromFile(*apiFile)
	if err != nil {
		return nil, fmt.Errorf("加载API配置文件失败: %w", err)
	}
	cfg.APIs = apis

	if *testDataFile != "" {
		testData, err := LoadTestData(*testDataFile)
		if err != nil {
			return nil, fmt.Errorf("加载测试数据失败: %w", err)
		}
		cfg.TestData = testData
	}

	return cfg, nil
}

// loadFromFile 和 loadAPIsFromFile 函数保持不变

func loadFromFile(filename string) (*Config, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = json.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func loadAPIsFromFile(filename string) (map[string]APIConfig, error) {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	var apis map[string]APIConfig
	err = json.Unmarshal(data, &apis)
	if err != nil {
		return nil, err
	}

	return apis, nil
}

func LoadTestData(filename string) ([]map[string]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	reader := csv.NewReader(file)
	records, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	headers := records[0]
	var testData []map[string]string
	for _, record := range records[1:] {
		data := make(map[string]string)
		for i, value := range record {
			data[headers[i]] = value
		}
		testData = append(testData, data)
	}

	return testData, nil
}
