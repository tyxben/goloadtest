package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

type LoginRequest struct {
	Type       string `json:"type"`
	WalletAddr string `json:"wallet_addr"`
	Text       string `json:"text"`
	Signature  string `json:"signature"`
}

type LoginResponse struct {
	Token string `json:"token"`
}

type UserInfo struct {
	WalletAddr       string         `json:"wallet_addr"`
	HasClaimedCarv   bool           `json:"has_claimed_carv"`
	HasClaimedVecarv bool           `json:"has_claimed_vecarv"`
	HasClaimedStake  bool           `json:"has_claimed_stake"`
	HasClaimedNft    bool           `json:"has_claimed_nft"`
	IsMagnaCarv      bool           `json:"is_magna_carv"`
	MagnaParameter   MagnaParameter `json:"magna_parameter,omitempty"`
	InviteInfo       InviteInfo     `json:"invite_info"`
}

type MagnaParameter struct {
	Data          string `json:"data"`
	To            string `json:"to"`
	From          string `json:"from"`
	Inputs        string `json:"inputs"`
	TransactionId string `json:"transactionId"`
}

type InviteInfo struct {
	InviteCode  string  `json:"invite_code"`
	InviteTotal int     `json:"invite_total"`
	EarnedCarv  float64 `json:"earned_carv"`
}

type StakingSpecialInfo struct {
	EstReward  string `json:"est_reward"`
	UnlockDate int64  `json:"unlock_date"`
}

type StakingSpecialSettings struct {
	Duration          int    `json:"duration"`
	StakingMultiplier int    `json:"staking_multiplier"`
	EstApr            string `json:"est_apr"`
}

type ApiResponse struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req LoginRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 这里应该有验证逻辑,但为了简化我们跳过了
	//时间戳 纳秒
	timestamp := time.Now().UnixNano()
	resp := LoginResponse{Token: "simulated_token_" + strconv.FormatInt(timestamp, 10)}
	json.NewEncoder(w).Encode(resp)
}

func userInfoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	auth := r.Header.Get("Authorization")
	if !strings.HasPrefix(auth, "simulated_token_") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	walletAddr := strings.TrimPrefix(auth, "simulated_token_")

	info := UserInfo{
		WalletAddr:       walletAddr,
		HasClaimedCarv:   false,
		HasClaimedVecarv: false,
		HasClaimedStake:  false,
		HasClaimedNft:    false,
		IsMagnaCarv:      true,
		MagnaParameter: MagnaParameter{
			Data:          "sample_data",
			To:            "sample_to",
			From:          "sample_from",
			Inputs:        "sample_inputs",
			TransactionId: "sample_transaction_id",
		},
		InviteInfo: InviteInfo{
			InviteCode:  "SAMPLE",
			InviteTotal: 30,
			EarnedCarv:  88.88,
		},
	}

	json.NewEncoder(w).Encode(info)
}

func stakingSpecialInfoHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 这里可以添加认证逻辑，如果需要的话

	// 获取查询参数
	walletAddr := r.URL.Query().Get("wallet_addr")
	fmt.Printf("walletAddr %s\n", walletAddr)
	amount := r.URL.Query().Get("amount")
	fmt.Printf("amount %s\n", amount)

	// 这里应该根据 walletAddr 和 amount 计算实际的奖励和解锁日期
	// 为了演示，我们使用固定值
	info := StakingSpecialInfo{
		EstReward:  "88.88",
		UnlockDate: time.Now().AddDate(0, 1, 0).Unix(), // 一个月后
	}

	response := ApiResponse{
		Code: 0,
		Msg:  "Success",
		Data: info,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func stakingSpecialSettingsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 这里可以添加认证逻辑，如果需要的话

	settings := StakingSpecialSettings{
		Duration:          2592000, // 30 天（以秒为单位）
		StakingMultiplier: 3,
		EstApr:            "120%",
	}

	response := ApiResponse{
		Code: 0,
		Msg:  "Success",
		Data: settings,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	http.HandleFunc("/airdrop/login", loginHandler)
	http.HandleFunc("/airdrop/user/info", userInfoHandler)

	http.HandleFunc("/explorer_testnet/staking_special_info", stakingSpecialInfoHandler)
	http.HandleFunc("/explorer_testnet/staking_special_settings", stakingSpecialSettingsHandler)

	fmt.Println("服务器正在启动,监听端口 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}
