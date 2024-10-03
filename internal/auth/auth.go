package auth

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/ethereum/go-ethereum/crypto"
)

// LoginResponse 定义登录响应的结构
type LoginResponse struct {
	Token string `json:"token"`
	// 可以根据实际登录响应添加其他字段
}

// Login 处理登录并获取token
func Login(loginURL string, loginBody []byte) (string, error) {
	resp, err := http.Post(loginURL, "application/json", bytes.NewBuffer(loginBody))
	if err != nil {
		return "", fmt.Errorf("登录请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("登录失败，状态码: %d", resp.StatusCode)
	}

	var loginResp LoginResponse
	if err := json.NewDecoder(resp.Body).Decode(&loginResp); err != nil {
		return "", fmt.Errorf("解析登录响应失败: %w", err)
	}

	if loginResp.Token == "" {
		return "", fmt.Errorf("登录响应中未找到token")
	}

	return loginResp.Token, nil
}

// SignMessage 使用私钥对消息进行签名
func SignMessage(message []byte, privateKeyHex string) (string, error) {
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return "", fmt.Errorf("解析私钥失败: %w", err)
	}

	hash := crypto.Keccak256Hash(message)
	signature, err := crypto.Sign(hash.Bytes(), privateKey)
	if err != nil {
		return "", fmt.Errorf("签名失败: %w", err)
	}

	return hexutil.Encode(signature), nil
}

// VerifySignature 验证签名
func VerifySignature(message []byte, signatureHex string, publicKeyHex string) (bool, error) {
	signature, err := hexutil.Decode(signatureHex)
	if err != nil {
		return false, fmt.Errorf("解析签名失败: %w", err)
	}

	publicKey, err := crypto.DecompressPubkey(hexutil.MustDecode(publicKeyHex))
	if err != nil {
		return false, fmt.Errorf("解析公钥失败: %w", err)
	}

	hash := crypto.Keccak256Hash(message)
	sigPublicKey, err := crypto.Ecrecover(hash.Bytes(), signature)
	if err != nil {
		return false, fmt.Errorf("恢复公钥失败: %w", err)
	}

	matches := bytes.Equal(crypto.FromECDSAPub(publicKey), sigPublicKey)
	return matches, nil
}

// GenerateNonce 生成一个基于时间戳的nonce
func GenerateNonce() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}

// SignRequest 对整个请求进行签名
func SignRequest(method, url string, body []byte, nonce string, privateKeyHex string) (string, error) {
	message := []byte(method + url + string(body) + nonce)
	return SignMessage(message, privateKeyHex)
}
