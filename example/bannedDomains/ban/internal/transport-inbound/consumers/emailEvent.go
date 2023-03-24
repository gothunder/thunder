package consumers

import (
	"context"
	"strings"

	thunderEvents "github.com/gothunder/thunder/pkg/events"
	"github.com/gothunder/thunder/example/email/pkg/events"
)

func (c *ConsumerGroup) emailEvent(ctx context.Context, payload events.EmailPayload) thunderEvents.HandlerResponse {
	domain := strings.Split(payload.Email, "@")[1]
	c.banService.CheckBan(ctx, domain, payload.ID)

	return thunderEvents.Success
}
