package main

import (
	"log"

	"github.com/tyxben/goloadtest/internal/runner"
	"github.com/tyxben/goloadtest/pkg/config"
)

func main() {
	cfg, err := config.Parse()
	if err != nil {
		log.Fatalf("解析配置失败: %v", err)
	}

	r := runner.NewRunner(cfg)
	r.Run()

	r.Stats.Print()
}
