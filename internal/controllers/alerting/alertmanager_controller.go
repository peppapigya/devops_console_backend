package alerting

import (
	"devops-console-backend/internal/common"
	alertingSvc "devops-console-backend/internal/services/alerting"
	"devops-console-backend/internal/types"
	"strings"

	"github.com/gin-gonic/gin"
)

type AlertmanagerController struct {
	orchestrator *alertingSvc.Orchestrator
}

func NewAlertmanagerController(orchestrator *alertingSvc.Orchestrator) *AlertmanagerController {
	return &AlertmanagerController{orchestrator: orchestrator}
}

func (ctrl *AlertmanagerController) Webhook(ctx *gin.Context) {
	if ctrl.orchestrator == nil || !ctrl.orchestrator.IsEnabled() {
		common.FailWithMsg(ctx, "alertmanager webhook is disabled")
		return
	}

	secret := strings.TrimSpace(ctrl.orchestrator.WebhookSecret())
	if secret != "" {
		headerValue := strings.TrimSpace(ctx.GetHeader(ctrl.orchestrator.AuthHeader()))
		if headerValue != secret {
			common.FailWithMsg(ctx, "invalid alertmanager token")
			return
		}
	}

	var payload types.AlertmanagerWebhookPayload
	if err := ctx.ShouldBindJSON(&payload); err != nil {
		common.FailWithMsg(ctx, "invalid alertmanager payload: "+err.Error())
		return
	}

	results, err := ctrl.orchestrator.HandleWebhook(payload)
	if err != nil {
		common.FailWithMsg(ctx, err.Error())
		return
	}

	common.Success(ctx, gin.H{
		"results": results,
		"count":   len(results),
	})
}
