package mocks

import (
	context "context"

	events "github.com/gothunder/thunder/pkg/events"
	. "github.com/onsi/ginkgo/v2"
	mock "github.com/stretchr/testify/mock"
)

type Handler struct {
	Mock mock.Mock
}

func (m *Handler) Handle(ctx context.Context, topic string, decoder events.EventDecoder) events.HandlerResponse {
	defer GinkgoRecover()

	args := m.Mock.Called(ctx, topic, decoder)

	var resp events.HandlerResponse
	if rf, ok := args.Get(0).(func(context.Context, string, events.EventDecoder) events.HandlerResponse); ok {
		resp = rf(ctx, topic, decoder)
	} else {
		resp = args.Get(0).(events.HandlerResponse)
	}

	return resp
}

func (m *Handler) Topics() []string {
	return []string{}
}

func (m *Handler) ResetMock() {
	m.Mock = mock.Mock{}
}

func NewHandler() *Handler {
	return &Handler{
		Mock: mock.Mock{},
	}
}
