package utiles

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type AlertConfig struct {
	WebhookURL string `json:"webhookUrl"` // Discord / Slack / 飞书 / 企业微信 webhook
	Enabled    bool   `json:"enabled"`
}

type AlertMessage struct {
	Level     string `json:"level"` // info / warning / error
	Title     string `json:"title"`
	Content   string `json:"content"`
	Timestamp string `json:"timestamp"`
}

// SendWebhookAlert 发送 webhook 通知（支持 Discord/Slack/飞书/企业微信/通用 webhook）
func SendWebhookAlert(webhookURL string, msg AlertMessage) error {
	if webhookURL == "" {
		return fmt.Errorf("webhook URL 为空")
	}

	msg.Timestamp = time.Now().Format("2006-01-02 15:04:05")

	// 自动识别 webhook 类型并构建对应 payload
	var payload []byte
	var err error
	var contentType string

	switch {
	case strings.Contains(webhookURL, "discord.com"):
		payload, err = buildDiscordPayload(msg)
		contentType = "application/json"
	case strings.Contains(webhookURL, "hooks.slack.com"):
		payload, err = buildSlackPayload(msg)
		contentType = "application/json"
	case strings.Contains(webhookURL, "feishu.cn") || strings.Contains(webhookURL, "larksuite.com"):
		payload, err = buildFeishuPayload(msg)
		contentType = "application/json"
	case strings.Contains(webhookURL, "qyapi.weixin.qq.com"):
		payload, err = buildWechatWorkPayload(msg)
		contentType = "application/json"
	default:
		// 通用 webhook，直接 POST JSON
		payload, err = json.Marshal(msg)
		contentType = "application/json"
	}
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Post(webhookURL, contentType, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("发送 webhook 失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return fmt.Errorf("webhook 返回错误状态: %d", resp.StatusCode)
	}
	return nil
}

func buildDiscordPayload(msg AlertMessage) ([]byte, error) {
	color := 0x00ff00 // green
	if msg.Level == "warning" {
		color = 0xffaa00
	} else if msg.Level == "error" {
		color = 0xff0000
	}
	payload := map[string]interface{}{
		"embeds": []map[string]interface{}{{
			"title":       msg.Title,
			"description": msg.Content,
			"color":       color,
			"footer":      map[string]string{"text": msg.Timestamp},
		}},
	}
	return json.Marshal(payload)
}

func buildSlackPayload(msg AlertMessage) ([]byte, error) {
	payload := map[string]interface{}{
		"text": fmt.Sprintf("*%s*\n%s\n_%s_", msg.Title, msg.Content, msg.Timestamp),
	}
	return json.Marshal(payload)
}

func buildFeishuPayload(msg AlertMessage) ([]byte, error) {
	payload := map[string]interface{}{
		"msg_type": "text",
		"content": map[string]string{
			"text": fmt.Sprintf("[%s] %s\n%s", msg.Level, msg.Title, msg.Content),
		},
	}
	return json.Marshal(payload)
}

func buildWechatWorkPayload(msg AlertMessage) ([]byte, error) {
	payload := map[string]interface{}{
		"msgtype": "text",
		"text": map[string]string{
			"content": fmt.Sprintf("[%s] %s\n%s", strings.ToUpper(msg.Level), msg.Title, msg.Content),
		},
	}
	return json.Marshal(payload)
}

// ContainerHealthAlert 容器健康检查告警
func ContainerHealthAlert(containerName, status string) AlertMessage {
	return AlertMessage{
		Level:   "warning",
		Title:   fmt.Sprintf("容器健康检查: %s", containerName),
		Content: fmt.Sprintf("容器 %s 状态变为 %s", containerName, status),
	}
}

// ContainerOOMAlert 容器 OOM 告警
func ContainerOOMAlert(containerName string) AlertMessage {
	return AlertMessage{
		Level:   "error",
		Title:   fmt.Sprintf("容器 OOM: %s", containerName),
		Content: fmt.Sprintf("容器 %s 因内存不足被杀死", containerName),
	}
}

// UpdateFailedAlert 更新失败告警
func UpdateFailedAlert(projectName, detail string) AlertMessage {
	return AlertMessage{
		Level:   "error",
		Title:   fmt.Sprintf("更新失败: %s", projectName),
		Content: detail,
	}
}
