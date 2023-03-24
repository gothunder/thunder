package consumers

import (
	thunderEvents "github.com/gothunder/thunder/pkg/events"
	"github.com/gothunder/thunder/example/ban/internal/features"
)

type ConsumerGroup struct {
	banService features.BanService
}

func newConsumerGroup(banService features.BanService) thunderEvents.Handler {
	return &ConsumerGroup{
		banService: banService,
	}
}
