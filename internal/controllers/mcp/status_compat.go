package mcp

import "devops-console-backend/internal/dal/model"

func normalizeSessionStatus(status string) string {
	switch status {
	case "pending":
		return "created"
	case "running":
		return "executing"
	case "waiting_confirm":
		return "waiting_approval"
	case "success", "partial":
		return "completed"
	default:
		return status
	}
}

func normalizeSessionForResponse(session *model.RepairSession) {
	if session == nil {
		return
	}
	session.Status = normalizeSessionStatus(session.Status)
}

func normalizeSessionListForResponse(list []model.RepairSession) {
	for i := range list {
		list[i].Status = normalizeSessionStatus(list[i].Status)
	}
}
