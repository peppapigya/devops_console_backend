package feishu

type Notifier interface {
	SendText(text string) error

	SendCard(title, content string) error

	SendChaosResult(experimentName, namespace, status string) error
}
