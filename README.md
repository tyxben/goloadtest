# GoLoadTest - 高性能 API 负载测试工具

GoLoadTest 是一个用 Go 语言编写的高性能 API 负载测试工具。它支持自定义工作流、并发测试和详细的统计报告，适用于各种 API 性能测试场景。

## 特点

- 支持自定义 API 工作流
- 可配置并发数和请求总数
- 支持从 CSV 文件读取测试数据
- 提供详细的测试统计报告
- 支持 HTTP 和 HTTPS 请求
- 支持自定义请求头和请求体
- 支持提取响应中的值用于后续请求

## 安装

确保您已安装 Go 1.16 或更高版本，然后运行以下命令：
```bash
git clone https://github.com/tysam/goloadtest.git
cd goloadtest
go build -o goloadtest cmd/main.go
```
## 配置文件

### config.json

此文件包含测试的全局配置。

```json
{
  "concurrency": 1,
  "duration": 1,
  "totalRequests" : 1,
  "workflow": ["login", "userInfo"],
  "tokenHeader": "Authorization",
  "baseURL": "http://localhost:8080"
}
```

### api.json

此文件定义了每个 API 的具体配置，包括所需的参数。

```json
{
  "login": {
    "url": "/airdrop/login",
    "method": "POST",
    "body": {
      "type": "wallet",
      "wallet_addr": "{{walletAddr}}",
      "text": "{{text}}",
      "signature": "{{signature}}"
    },
    "response": {
      "token": "token"
    },
    "params": ["walletAddr", "text", "signature"]
  },
  "userInfo": {
    "url": "/airdrop/user/info",
    "method": "GET",
    "headers": {
      "Authorization": "{{token}}"
    },
    "params": []
  },
  "stakingSpecialInfo": {
    "url": "/explorer_testnet/staking_special_info",
    "method": "GET",
    "params": ["walletAddr", "amount"],
    "queryParams": {
      "wallet_addr": "{{walletAddr}}",
      "amount": "{{amount}}"
    }
  }
}
```

每个 API 配置中的 `params` 字段定义了该 API 所需的参数列表。这些参数将从测试数据中读取。

## 测试数据

测试数据可以通过 CSV 文件提供，支持多个参数。例如 `testdata.csv`：

```csv
walletAddr,amount,text,signature
0x1111111111111111111111111111111111111111,100,test_text_1,sig_1
0x2222222222222222222222222222222222222222,200,test_text_2,sig_2
0x3333333333333333333333333333333333333333,300,test_text_3,sig_3
```

CSV 文件应包含所有 API 可能用到的参数。每个 API 只会使用其配置中定义的参数。

## 使用方法

运行测试时，指定配置文件和测试数据文件：

```bash
./goloadtest -config config.json -api api.json -testdata testdata.csv
```

## 扩展性

1. 动态参数：在 `api.json` 中，使用 `{{paramName}}` 语法可以引用测试数据中的任何列。
2. 自定义数据源：除了 CSV 文件，您还可以实现其他数据源，如数据库或 API。
3. 参数转换：在 `internal/worker/worker.go` 中，您可以添加函数来处理和转换参数，例如生成动态签名。

## 注意事项

1. 确保 CSV 文件中包含所有 API 可能用到的参数。
2. 如果某个 API 不需要特定参数，可以在 CSV 文件中留空，或在代码中处理缺失参数的情况。
3. 对于大规模测试，考虑使用数据库或其他高效的数据源来提供测试数据。

## 故障排除

- 如果遇到 "connection refused" 错误，请检查目标服务器是否正在运行，以及 `baseURL` 是否配置正确
- 如果看到 "invalid character" 错误，请检查 JSON 配置文件的格式是否正确
- 如果测试数据不生效，确保 CSV 文件的路径正确，且文件格式符合要求

## 性能建议

- 增加并发数可以提高吞吐量，但也会增加目标服务器的负载
- 使用 SSD 存储可以提高大量并发请求的性能
- 如果测试大规模并发，考虑使用多台机器分布式运行测试

## 贡献

欢迎提交 Issue 和 Pull Request 来改进这个项目。在提交 PR 之前，请确保您的代码符合 Go 的代码规范，并且通过了所有的测试。

## 联系方式

如果您有任何问题或建议，请通过以下方式联系我们：

- 提交 GitHub Issue

