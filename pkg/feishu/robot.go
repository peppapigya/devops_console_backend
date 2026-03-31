package feishu

type MessageType string

const (
	TextMessageType MessageType = "text"
)

type TextMessage struct {
	MsgType MessageType `json:"msg_type"`
	Content struct {
		Text string `json:"text"`
	} `json:"content"`
}

// SendText 发送文本消息
func (c *Client) SendText(text string) error {
	message := &TextMessage{
		MsgType: TextMessageType,
	}

	message.Content.Text = text
	return c.post(message)
}

// SendCard 发送卡片消息
func (c *Client) SendCard(title, content string) error {
	body := map[string]interface{}{
		"msg_type": "interactive",
		"card": map[string]interface{}{
			"header": map[string]interface{}{
				"title": map[string]interface{}{
					"tag":     "plain_text",
					"content": title,
				},
			},
			"elements": []map[string]interface{}{
				{
					"tag": "div",
					"text": map[string]interface{}{
						"tag":     "lark_md",
						"content": content,
					},
				},
			},
		},
	}
	return c.post(body)

}

func (c *Client) SendChaosResult(experimentName, namespace, status string) error {
	return nil
}
