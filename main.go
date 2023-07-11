package main

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"time"
)

var (
	DefaultPasswordEncoded = "3fdebd70bc927d97"
	DefaultPassword        = "zjucst"
)

func main() {
	// 在环境变量中填写：
	fmt.Println("使用前请在环境变量中填写配置 CST_USERNAME_CONFIG：")
	fmt.Println("[WIFI_NAME1]:[Username1]【,[Password1]】;[WIFI_NAME2]:[Username2]【,[Password2]】;")
	fmt.Println("请注意Password（可选）为编码后的字符串，请使用网页开发者工具查看")
	fmt.Println("--------------------------------------------------")

	wifiMap := map[string][2]string{}

	// 1.获取环境变量配置CST_USERNAME_CONFIG -> wifiMap
	// rawVal: [WIFI_NAME1]:[Username1]【,[Password1]】;[WIFI_NAME2]:[Username2]【,[Password2]】;
	if rawVal := os.Getenv("CST_USERNAME_CONFIG"); rawVal != "" {
		// 去除前后;
		rawVal = strings.Trim(rawVal, ";")
		// spilt rawVal ;
		vals := strings.Split(rawVal, ";")
		for _, val := range vals {
			item := strings.Split(val, ":")
			if len(item) != 2 {
				fmt.Println("环境变量格式错误：CST_USERNAME_CONFIG, Value格式为：[WIFI_NAME1]:[Username1]【,[Password1]】;[WIFI_NAME2]:[Username2]【,[Password2]】;")
				ScanToExit()
			}

			wifiName := item[0]
			username := item[1]
			password := DefaultPasswordEncoded
			if strings.Contains(username, ",") {
				pp := strings.Split(username, ",")
				username = pp[0]
				password = pp[1]
			}
			wifiMap[wifiName] = [2]string{username, password}
		}
	} else {
		fmt.Println("未设置环境变量：CST_USERNAME_CONFIG, Value格式为：[WIFI_NAME1]:[Username1]【,[Password1]】;[WIFI_NAME2]:[Username2]【,[Password2]】;")
		ScanToExit()
	}

	// 2.检测并获取WIFI-name
	targetUsername := ""
	targetPassword := ""

	for {
		fmt.Println("等待目标WiFi连接...")
		wifiName := getCurrentWifiName()
		//fmt.Println(wifiName)
		userNameAndPwd, exists := wifiMap[wifiName]
		if exists {
			fmt.Printf("匹配到WiFi Name: %v, Username: %v\n", wifiName, userNameAndPwd[0])
			targetUsername = userNameAndPwd[0]
			targetPassword = userNameAndPwd[1]
			break
		}
		time.Sleep(1000)
	}

	fmt.Println("开始登录...")

	maxRetryTimes := 3
	retryTimes := 0
	connected := false
	// 登录
	for retryTimes < maxRetryTimes {
		ok := Login(targetUsername, targetPassword)
		time.Sleep(2 * time.Second)
		// 检测是否联网
		if ok && isInternetConnected() {
			connected = true
			break
		}
		retryTimes++
		fmt.Printf("联网检测失败，正在重试(%d/%d)\n", retryTimes, maxRetryTimes)
	}

	if !connected {
		fmt.Println("登录失败")
	} else {
		fmt.Println("登录成功")
	}

	time.Sleep(2 * time.Second)
}

func Login(username, decodedPwd string) bool {
	// 构造请求体
	data := url.Values{}
	data.Set("username", username)
	data.Set("password", decodedPwd)
	data.Set("drop", "0")
	data.Set("type", "1")
	data.Set("n", "100")
	requestBody := bytes.NewBufferString(data.Encode())

	// 构造请求头
	headers := map[string]string{
		"Accept":          "*/*",
		"Accept-Encoding": "gzip, deflate",
		"Accept-Language": "zh-CN,zh;q=0.9",
		"Connection":      "keep-alive",
		"Content-Length":  fmt.Sprintf("%d", requestBody.Len()),
		"Content-Type":    "application/x-www-form-urlencoded",
		//"Cookie":          fmt.Sprintf("srun_login=%v", username) + "%7C" + Password,
		"Host":       "192.0.0.6",
		"Origin":     "http://192.0.0.6",
		"Referer":    "http://192.0.0.6/",
		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/103.0.0.0 Safari/537.36",
	}

	// 构造请求对象
	req, err := http.NewRequest("POST", "http://192.0.0.6/cgi-bin/do_login", requestBody)
	if err != nil {
		fmt.Println("构造请求失败：", err)
		return false
	}
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	// 发送请求
	client := &http.Client{
		Transport: &http.Transport{}, // 强制走系统代理
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("发送请求失败：", err)
		return false
	}
	defer resp.Body.Close()

	// 处理响应
	fmt.Println(resp.Status)
	return strings.Contains(resp.Status, "200")
}

func isInternetConnected() bool {
	_, err := http.Get("http://www.baidu.com")
	if err != nil {
		return false
	}
	return true
}

func getCurrentWifiName() string {
	cmd := exec.Command("cmd", "/C", "netsh", "wlan", "show", "interfaces")
	var out bytes.Buffer
	cmd.Stdout = &out
	err := cmd.Run()

	if err != nil {
		fmt.Println("命令netsh执行失败")
		ScanToExit()
	}

	for _, line := range strings.Split(out.String(), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "SSID") && !strings.Contains(line, "BSSID") {
			ssid := strings.Split(line, ":")
			if len(ssid) > 1 {
				return strings.TrimSpace(ssid[1])
			}
		}
	}

	return ""
}

func ScanToExit() {
	s := ""
	fmt.Scanf("%s", &s)
	os.Exit(0)
}
