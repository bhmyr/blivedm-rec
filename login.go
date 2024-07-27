package main

import (
	"bytes"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"regexp"
	"time"
)

type LoginInfo struct {
	URL          string            `json:"url"`
	RefreshToken string            `json:"refresh_token"`
	Cookies      map[string]string `json:"cookies"`
}

var loginInfo *LoginInfo

// 获取二维码登录链接，返回二维码链接和二维码key
func (l *LoginInfo) get_qrcode() (string, string) {
	slog.Info("获取二维码")
	url := "https://passport.bilibili.com/x/passport-login/web/qrcode/generate"

	var qrcode struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		TTL     int    `json:"ttl"`
		Data    struct {
			URL       string `json:"url"`
			QrcodeKey string `json:"qrcode_key"`
		} `json:"data"`
	}
	resp := request_get(url, nil)

	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&qrcode); err == nil {
		if qrcode.Code != 0 {
			log.Fatal("Error: QR code generation failed")
		}
	} else {
		log.Fatal("Error: QR code generation failed")
	}

	return qrcode.Data.URL, qrcode.Data.QrcodeKey
}

// 登录
func (l *LoginInfo) Login() {
	qr_url, qr_token := l.get_qrcode()

	fmt.Println("在浏览器打开链接，并扫码登录: " + qr_url)
	fmt.Print("扫码登录成功后按回车键继续...")
	fmt.Scanln()
	var data struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Data    struct {
			URL          string `json:"url"`
			RefreshToken string `json:"refresh_token"`
			Timestamp    int    `json:"timestamp"`
			Code         int    `json:"code"`
			Message      string `json:"message"`
		} `json:"data"`
	}
	resp := request_get("https://passport.bilibili.com/x/passport-login/web/qrcode/poll?qrcode_key="+qr_token, nil)

	defer resp.Body.Close()
	if err := json.NewDecoder(resp.Body).Decode(&data); err == nil {
		if data.Data.RefreshToken == "" {
			log.Fatal("Error: " + data.Data.Message)
		}
		l.URL = data.Data.URL
		l.RefreshToken = data.Data.RefreshToken
		for _, c := range resp.Cookies() {
			l.Cookies[c.Name] = c.Value
		}
	} else {
		log.Fatal("Error: " + data.Message)
	}
	struct2json(*l)
}

// 是否需要刷新cookie
func (l *LoginInfo) refresh_check() bool {
	slog.Info("检查是否需要刷新")
	url := "https://passport.bilibili.com/x/passport-login/web/cookie/info"
	resp := request_get(url, l.Cookies)
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return false
	}
	var data struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		TTL     int    `json:"ttl"`
		Data    struct {
			Refresh   bool  `json:"refresh"`
			Timestamp int64 `json:"timestamp"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return false
	}
	return true
}

func (l *LoginInfo) get_refresh_csrf(CorrespondPath string) string {
	slog.Info("获取刷新csrf")
	url := "https://www.bilibili.com/correspond/1/" + CorrespondPath
	resp := request_get(url, l.Cookies)
	defer resp.Body.Close()

	b, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	data := string(b)

	re := regexp.MustCompile(`id="1-name">(.*?)<\/div>`)
	matches := re.FindStringSubmatch(data)
	if len(matches) > 1 {
		content := matches[1]
		return content
	} else {
		return ""
	}

}

func (l *LoginInfo) Refresh_cookie() {

	if !l.refresh_check() {
		return
	}
	ts := time.Now().UnixMicro()
	CorrespondPath, _ := getCorrespondPath(ts)
	refresh_csrf := l.get_refresh_csrf(CorrespondPath)
	if refresh_csrf == "" {
		return
	}
	crsf := get_csrf(l.Cookies)
	//刷新
	slog.Info("刷新cookie")
	url := "https://passport.bilibili.com/x/passport-login/web/cookie/refresh"

	jsonData := map[string]interface{}{
		"csrf":          crsf,
		"refresh_csrf":  refresh_csrf,
		"source":        "main_web",
		"refresh_token": l.RefreshToken,
	}
	resp := request_post(url, l.Cookies, jsonData)
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Fatal("Error: Failed to refresh token")
	}
	var data struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		TTL     int    `json:"ttl"`
		Data    struct {
			Status       int    `json:"status"`
			Message      string `json:"message"`
			RefreshToken string `json:"refresh_token"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		log.Fatal("Error: Failed to refresh token")
	}
	cookie_new := map[string]string{}
	for _, c := range resp.Cookies() {
		cookie_new[c.Name] = c.Value
	}

	if l.confirm_refresh(cookie_new) {
		slog.Info("刷新成功，更新cookie")
		l.Cookies = cookie_new
		l.RefreshToken = data.Data.RefreshToken
		struct2json(*l)
	}
}

// 确认刷新
func (l *LoginInfo) confirm_refresh(cookie map[string]string) bool {
	slog.Info("确认刷新")
	url := "https://passport.bilibili.com/x/passport-login/web/confirm/refresh"
	crsf := get_csrf(cookie)
	jsonData := map[string]interface{}{
		"csrf":          crsf,
		"refresh_token": l.RefreshToken,
	}
	var data struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		TTL     int    `json:"ttl"`
	}
	resp := request_post(url, cookie, jsonData)
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		slog.Info("Error: Failed to refresh token")
	}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		slog.Info("Error: Failed to refresh token")
	}
	if data.Code != 0 {

		slog.Info("Error: Failed to refresh token " + data.Message)
		return false
	}
	return true
}

func get_csrf(cookies map[string]string) string {
	crsf := ""
	for k, c := range cookies {
		if k == "bili_jct" {
			crsf = c
			break
		}
	}
	return crsf
}

func map2cookie(cookie map[string]string) *http.Cookie {
	var cookies *http.Cookie
	for k, v := range cookie {
		cookies = &http.Cookie{
			Name:  k,
			Value: v,
		}
	}
	return cookies
}

func request_get(url string, cookie map[string]string) *http.Response {
	client := &http.Client{}
	req, _ := http.NewRequest("GET", url, nil)
	if cookie != nil {
		req.AddCookie(map2cookie(cookie))
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Error: Request failed")
	}
	return resp
}

func request_post(url string, cookie map[string]string, data map[string]interface{}) *http.Response {
	client := &http.Client{}
	jsonData, _ := json.Marshal(data)
	req, _ := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if cookie != nil {
		req.AddCookie(map2cookie(cookie))
	}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal("Error: Request failed")
	}
	return resp
}

// https://github.com/SocialSisterYi/bilibili-API-collect/blob/master/docs/login/cookie_refresh.md#Python
func getCorrespondPath(ts int64) (string, error) {
	const publicKeyPEM = `
-----BEGIN PUBLIC KEY-----
MIGfMA0GCSqGSIb3DQEBAQUAA4GNADCBiQKBgQDLgd2OAkcGVtoE3ThUREbio0Eg
Uc/prcajMKXvkCKFCWhJYJcLkcM2DKKcSeFpD/j6Boy538YXnR6VhcuUJOhH2x71
nzPjfdTcqMz7djHum0qSZA0AyCBDABUqCrfNgCiJ00Ra7GmRj+YCK1NJEuewlb40
JNrRuoEUXpabUzGB8QIDAQAB
-----END PUBLIC KEY-----
`
	pubKeyBlock, _ := pem.Decode([]byte(publicKeyPEM))
	hash := sha256.New()
	random := rand.Reader
	msg := []byte(fmt.Sprintf("refresh_%d", ts))
	var pub *rsa.PublicKey
	pubInterface, parseErr := x509.ParsePKIXPublicKey(pubKeyBlock.Bytes)
	if parseErr != nil {
		return "", parseErr
	}
	pub = pubInterface.(*rsa.PublicKey)
	encryptedData, encryptErr := rsa.EncryptOAEP(hash, random, pub, msg, nil)
	if encryptErr != nil {
		return "", encryptErr
	}
	return hex.EncodeToString(encryptedData), nil
}

func struct2json(data LoginInfo) {
	jsonData, _ := json.Marshal(data)
	os.WriteFile("login.json", jsonData, 0644)
}

func login() {
	loginInfo = &LoginInfo{
		Cookies: map[string]string{},
	}
	jsonData, err := os.ReadFile("login.json")

	if err != nil {
		loginInfo.Login()
	} else {
		json.Unmarshal(jsonData, &loginInfo)
	}
	go func() {
		for {
			loginInfo.Refresh_cookie()
			time.Sleep(5 * time.Hour)
		}
	}()
}
