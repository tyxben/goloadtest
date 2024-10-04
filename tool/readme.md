安装必要的依赖：
ethereum
编译程序：
go
使用方法：
生成钱包：
  ./wallet_tool generate -count 20 -output wallets.csv
签名消息：
  ./wallet_tool sign -input wallets.csv -output ../test.csv

这个工具现在有两个主要功能：
generate 命令：生成指定数量的钱包地址和私钥，并保存到 CSV 文件中。
sign 命令：读取包含地址和私钥的 CSV 文件，为每个地址生成一个签名消息，并将结果保存到新的 CSV 文件中。