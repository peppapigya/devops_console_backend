package event

import (
	"devops-console-backend/internal/controllers/k8s/event"

	"github.com/gin-gonic/gin"
)

type EventRoute struct {
	eventController *event.EventController
}

func NewEventRoute() *EventRoute {
	return &EventRoute{
		eventController: event.NewEventController(),
	}
}

func (r *EventRoute) RegisterSubRouter(apiGroup *gin.RouterGroup) {
	// Event
	eventGroup := apiGroup.Group("/k8s/event")
	{
		eventGroup.GET("/list/:namespace", r.eventController.GetEventList)
		eventGroup.GET("/list/all", r.eventController.GetEventList)
	}
}
