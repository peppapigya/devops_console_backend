package feishu

import (
	"devops-console-backend/pkg/utils/logs"
	"testing"
)

const WebhookUrl = "https://open.feishu.cn/open-apis/bot/v2/hook/xxxxxxxxxxxx"

func TestFeiShuSendText(t *testing.T) {
	client := NewClient(WebhookUrl, "", "", "", true)

	err := client.SendText("hello world")
	if err != nil {
		logs.Error(map[string]interface{}{"error": err.Error()}, "发送飞书文本消息失败")
		return
	}
}

func TestFeiShuSendCard(t *testing.T) {
	client := NewClient(WebhookUrl, "", "", "", true)

	err := client.SendCard("这是标题", "这是内容")
	if err != nil {
		logs.Error(map[string]interface{}{"error": err.Error()}, "发送飞书卡片消息失败")
		return
	}
}
