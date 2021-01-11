package pubsubmock

import (
	"context"
	"sync"

	"github.com/golang/protobuf/proto"
	log "github.com/sirupsen/logrus"
)

type pubsubmock struct {
	mu     sync.Mutex
	events []proto.Message
}

// New retruns pubsubmock implementation.
// It should be replaced with real pubsub in future.
// It is used only development purpose.
func New() *pubsubmock {
	return &pubsubmock{}
}

func (p *pubsubmock) Publish(_ context.Context, in proto.Message) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.events = append(p.events, in)
	log.WithField("msg", in).Infof("Received event: %T", in)
	return nil
}
