package feishu

import (
	"bytes"
	"devops-console-backend/pkg/utils/logs"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client 飞书客户端
type Client struct {
	webHookUrl *string
	secret     string
	appID      string
	appSecret  string
	enabled    bool
	HttpClient *http.Client
}

func NewClient(
	webHookUrl string,
	secret string,
	appID string,
	appSecret string,
	enabled bool,
) *Client {
	return &Client{
		webHookUrl: &webHookUrl,
		secret:     secret,
		appID:      appID,
		appSecret:  appSecret,
		enabled:    enabled,
		HttpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *Client) post(body interface{}) error {
	if !c.enabled {
		return nil
	}

	if c.webHookUrl == nil || *c.webHookUrl == "" {
		return fmt.Errorf("webhook url is empty")
	}

	// 反解析
	data, err := json.Marshal(body)
	if err != nil {
		logs.Error(map[string]interface{}{"error": err.Error()}, "json marshal err")
		return err
	}

	// create request
	request, err := http.NewRequest(http.MethodPost, *c.webHookUrl, bytes.NewReader(data))
	if err != nil {
		logs.Error(map[string]interface{}{"error": err.Error()}, "create request err")
		return err
	}

	//设置请求头
	request.Header.Set("Content-Type", "application/json")
	resp, err := c.HttpClient.Do(request)
	if err != nil {
		logs.Error(map[string]interface{}{"error": err.Error()}, "http request err")
		return err
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("http status code: %d", resp.StatusCode)
	}

	respBody, _ := io.ReadAll(resp.Body)
	logs.Info(map[string]interface{}{"resp": string(respBody)}, "飞书消息响应内容")
	return nil

}
