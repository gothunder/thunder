package transportoutbound

import (
	"github.com/gothunder/thunder/example/email/internal/transport-outbound/publisher"
	"go.uber.org/fx"
)

var Module = fx.Options(
	publisher.Module,
)
