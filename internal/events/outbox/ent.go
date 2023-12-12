package outbox

import (
	"context"
	"errors"
	"reflect"

	"github.com/TheRafaBonin/roxy"
	"github.com/gothunder/thunder/internal/utils"
)

const (
	SetTopic   = "SetTopic"
	SetHeaders = "SetHeaders"
	SetPayload = "SetPayload"
	Exec       = "Exec"

	Create     = "Create"
	CreateBulk = "CreateBulk"
)

var (
	ErrNilClient = errors.New("nil OutboxMessageClient")
)

type MessageClient interface {
	Create() MessageCreator
	CreateBulk(messageCreators ...MessageCreator) MessageBulkCreator
}

type MessageCreator interface {
	SetTopic(topic string) MessageCreator
	SetPayload(payload []byte) MessageCreator
	SetHeaders(headers map[string]string) MessageCreator
	Exec(ctx context.Context) error
	Unwrap() interface{}
}

type MessageBulkCreator interface {
	Exec(ctx context.Context) error
}

type outboxMessageCreateWrapper struct {
	outboxMessageCreate interface{}
}

func (omcw *outboxMessageCreateWrapper) SetHeaders(headers map[string]string) MessageCreator {
	_, err := utils.SafeCallMethod(omcw.outboxMessageCreate, SetHeaders, []reflect.Value{
		reflect.ValueOf(headers),
	})
	if err != nil {
		panic(err)
	}
	return omcw
}

func (omcw *outboxMessageCreateWrapper) SetPayload(payload []byte) MessageCreator {
	_, err := utils.SafeCallMethod(omcw.outboxMessageCreate, SetPayload, []reflect.Value{
		reflect.ValueOf(payload),
	})
	if err != nil {
		panic(err)
	}
	return omcw
}

func (omcw *outboxMessageCreateWrapper) SetTopic(topic string) MessageCreator {
	_, err := utils.SafeCallMethod(omcw.outboxMessageCreate, SetTopic, []reflect.Value{
		reflect.ValueOf(topic),
	})
	if err != nil {
		panic(err)
	}
	return omcw
}

func (omcw *outboxMessageCreateWrapper) Exec(ctx context.Context) error {
	results, err := utils.SafeCallMethod(omcw.outboxMessageCreate, Exec, []reflect.Value{
		reflect.ValueOf(ctx),
	})
	if err != nil {
		return err
	}

	return results[0].Interface().(error)
}

func (omcw *outboxMessageCreateWrapper) Unwrap() interface{} {
	return omcw.outboxMessageCreate
}

func WrapOutboxMessageCreate(omc interface{}) (MessageCreator, error) {
	handleError := func(methodName string) (MessageCreator, error) {
		return nil, roxy.Wrap(utils.ErrMethodNotFound, methodName)
	}

	methods := []string{SetTopic, SetHeaders, SetPayload, Exec}
	for _, method := range methods {
		if !utils.HasMethod(omc, method) {
			return handleError(method)
		}
	}

	return &outboxMessageCreateWrapper{outboxMessageCreate: omc}, nil
}

type MessageBuilderInterface[T any] interface {
	SetHeaders(headers map[string]string) T
	SetPayload(payload []byte) T
	SetTopic(topic string) T
	Exec(ctx context.Context) error
}

type outboxMessageClientWrapper struct {
	OutboxMessageClient interface{}
}

func (c *outboxMessageClientWrapper) Create() MessageCreator {
	results, err := utils.SafeCallMethod(c.OutboxMessageClient, Create, []reflect.Value{})
	if err != nil {
		panic(err)
	}
	creator, _ := WrapOutboxMessageCreate(results[0].Interface())
	return creator
}

func (c *outboxMessageClientWrapper) CreateBulk(messageCreators ...MessageCreator) MessageBulkCreator {
	creators := make([]reflect.Value, len(messageCreators))
	for i, mc := range messageCreators {
		creators[i] = reflect.ValueOf(mc.Unwrap())
	}

	results, err := utils.SafeCallMethod(c.OutboxMessageClient, CreateBulk, creators)
	if err != nil {
		panic(err)
	}

	return results[0].Interface().(MessageBulkCreator)
}

func WrapOutboxMessageClient(client interface{}) (MessageClient, error) {
	if err := validateOutboxMessageClient(client); err != nil {
		return nil, roxy.Wrap(err, "validating client")
	}

	return &outboxMessageClientWrapper{OutboxMessageClient: client}, nil
}

func validateOutboxMessageClient(client interface{}) error {
	if client == nil {
		return ErrNilClient
	}

	handleError := func(methodName string) error {
		return roxy.Wrap(utils.ErrMethodNotFound, methodName)
	}

	methods := []string{Create, CreateBulk}
	for _, method := range methods {
		if !utils.HasMethod(client, method) {
			return handleError(method)
		}
	}

	return nil
}

type OutboxMessageClientInterface[T MessageBuilderInterface[T], U MessageBulkCreator] interface {
	Create() T
	CreateBulk(messageCreators ...T) U
}
