package outboxent

import (
	"context"
	"errors"
	"reflect"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/TheRafaBonin/roxy"
	"github.com/google/uuid"
	"github.com/gothunder/thunder/internal/events/outbox"
	"github.com/gothunder/thunder/internal/utils"
	"github.com/rs/zerolog"
)

var (
	ErrMessagePollerClosed = errors.New("message poller closed")
	ErrInvalidMessagePack  = errors.New("invalid message pack")
	ErrInvalidPollInterval = errors.New("invalid poll interval")
	ErrInvalidBatchSize    = errors.New("invalid batch size")
)

const (
	// methods
	methodQuery      = "Query"
	methodQueryWhere = "Where"
	methodQueryAll   = "All"
	methodLimit      = "Limit"
	methodOrder      = "Order"

	// fields
	fieldCreatedAt = "created_at"
)

func NewEntMessagePoller(
	outboxMessageClient interface{},
	pollInterval time.Duration,
	batchSize int,
) (outbox.MessagePoller, error) {
	if outboxMessageClient == nil || !utils.HasMethod(outboxMessageClient, methodQuery) {
		return nil, roxy.Wrap(utils.ErrMethodNotFound, methodQuery)
	}
	if pollInterval <= 0 {
		return nil, ErrInvalidPollInterval
	}
	if batchSize <= 0 {
		return nil, ErrInvalidBatchSize
	}

	return &entMessagePoller{
		client:       outboxMessageClient,
		pollInterval: pollInterval,
		batchSize:    batchSize,
		closeChan:    make(chan struct{}),
		nextChan:     make(chan struct{}),
		closed:       false,
	}, nil
}

type entMessagePoller struct {
	client interface{}

	pollInterval time.Duration
	batchSize    int

	closed    bool
	closeChan chan struct{}
	// nextChan is used to notify the poller to load more messages.
	// It is used to avoid duplicate messages due to eaguer loading
	// of one batch until the message channel blocks awaiting consumer.
	// This way it awaits the consumer to tell it already processed that
	// batch of messages.
	nextChan chan struct{}
}

// Close implements outbox.MessagePoller.
func (e *entMessagePoller) Close() error {
	if e.closed {
		return nil
	}
	close(e.closeChan)
	close(e.nextChan)
	e.closed = true
	return nil
}

// Poll implements outbox.MessagePoller.
func (e *entMessagePoller) Poll(ctx context.Context) (<-chan []*outbox.Message, func(), error) {
	if e.closed {
		return nil, func() {}, ErrMessagePollerClosed
	}

	messageChan := make(chan []*outbox.Message)
	logger := zerolog.Ctx(ctx)

	go func() {
		defer close(messageChan)

		for {
			select {
			case <-ctx.Done():
				return
			case <-e.closeChan:
				return
			default:
				if err := e.forwardMessages(ctx, messageChan); err != nil {
					if errors.Is(err, ErrMessagePollerClosed) || errors.Is(err, context.Canceled) {
						return
					}
					logger.Err(err).Msg("error polling outbox messages")
					panic(err)
				}

				time.Sleep(e.pollInterval)
			}
		}
	}()

	return messageChan, func() {
		if e.closed {
			return
		}
		e.nextChan <- struct{}{}
	}, nil
}

func (e *entMessagePoller) forwardMessages(ctx context.Context, msgChan chan<- []*outbox.Message) error {
	var err error
	var msgPack []*outbox.Message
	for msgPack, err = e.readBatch(ctx); len(msgPack) > 0 && err == nil; msgPack, err = e.readBatch(ctx) {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-e.closeChan:
			return ErrMessagePollerClosed
		default:
			msgChan <- msgPack
			// await notification to load more messages
			<-e.nextChan
		}
	}

	return roxy.Wrap(err, "reading batch of messages")
}

func (e *entMessagePoller) readBatch(ctx context.Context) ([]*outbox.Message, error) {
	queryBuilder, err := newQueryBuilder(e.client)
	if err != nil {
		return nil, roxy.Wrap(err, "creating query builder")
	}

	if err := queryBuilder.WhereDeliveryAtIsNil(); err != nil {
		return nil, roxy.Wrap(err, "setting where clause")
	}

	if err := queryBuilder.Limit(e.batchSize); err != nil {
		return nil, roxy.Wrap(err, "setting limit clause")
	}

	if err := queryBuilder.OrderByCreatedAt(); err != nil {
		return nil, roxy.Wrap(err, "setting order clause")
	}

	return queryBuilder.All(ctx)
}

type outboxMessageQueryBuilder struct {
	builder interface{}
}

func newQueryBuilder(client interface{}) (*outboxMessageQueryBuilder, error) {
	// initialize query builder
	queryBuilder, err := utils.SafeCallMethod(client, methodQuery, []reflect.Value{})
	if err != nil {
		return nil, roxy.Wrap(err, "calling query method on ent client")
	}

	return &outboxMessageQueryBuilder{
		builder: queryBuilder[0].Interface(),
	}, nil
}

func (q *outboxMessageQueryBuilder) WhereDeliveryAtIsNil() error {
	method, ok := reflect.TypeOf(q.builder).MethodByName(methodQueryWhere)
	if !ok {
		return roxy.Wrap(utils.ErrMethodNotFound, "getting method Where of OutboxMessageQuery")
	}

	elemType := method.Type.In(1).Elem()
	deliveredAtNilClause := reflect.ValueOf(sql.FieldIsNull(fieldDeliveredAt)).Convert(elemType)

	_, err := utils.SafeCallMethod(q.builder, methodQueryWhere, []reflect.Value{
		deliveredAtNilClause,
	})
	if err != nil {
		return roxy.Wrap(err, "calling Where method on OutboxMessageQuery")
	}
	return nil
}

func (q *outboxMessageQueryBuilder) Limit(limit int) error {
	_, err := utils.SafeCallMethod(q.builder, methodLimit, []reflect.Value{
		reflect.ValueOf(limit),
	})
	if err != nil {
		return roxy.Wrap(err, "calling Limit method on OutboxMessageQuery")
	}

	return nil
}

func (q *outboxMessageQueryBuilder) OrderByCreatedAt() error {
	method, ok := reflect.TypeOf(q.builder).MethodByName(methodOrder)
	if !ok {
		return roxy.Wrap(utils.ErrMethodNotFound, "getting method Order of OutboxMessageQuery")
	}

	elemType := method.Type.In(1).Elem()
	orderClause := reflect.ValueOf(sql.OrderByField(fieldCreatedAt, sql.OrderAsc()).ToFunc()).Convert(elemType)

	_, err := utils.SafeCallMethod(q.builder, methodOrder, []reflect.Value{
		orderClause,
	})
	if err != nil {
		return roxy.Wrap(err, "calling Order method on OutboxMessageQuery")
	}

	return nil
}

func (q *outboxMessageQueryBuilder) All(ctx context.Context) ([]*outbox.Message, error) {
	// All
	result, err := utils.SafeCallMethod(q.builder, methodQueryAll, []reflect.Value{
		reflect.ValueOf(ctx),
	})
	if err != nil {
		return nil, roxy.Wrap(err, "calling All method on OutboxMessageQuery")
	}
	if result[1].Interface() != nil {
		return nil, roxy.Wrap(result[1].Interface().(error), "getting result of All method on OutboxMessageQuery")
	}

	return parseEntMessages(result[0])
}

func parseEntMessages(result reflect.Value) ([]*outbox.Message, error) {
	msgPack := result
	if msgPack.Kind() != reflect.Slice {
		return nil, roxy.Wrap(ErrInvalidMessagePack, "getting slice of messages")
	}

	messages := make([]*outbox.Message, msgPack.Len())
	for i := 0; i < msgPack.Len(); i++ {
		msg := msgPack.Index(i)
		if msg.Kind() == reflect.Ptr {
			msg = msg.Elem()
		}
		if msg.Kind() != reflect.Struct {
			return nil, roxy.Wrap(ErrInvalidMessagePack, "getting message struct")
		}

		messages[i] = &outbox.Message{
			ID:        msg.FieldByName("ID").Interface().(uuid.UUID),
			CreatedAt: msg.FieldByName("CreatedAt").Interface().(time.Time),
			Topic:     msg.FieldByName("Topic").Interface().(string),
			Payload:   msg.FieldByName("Payload").Interface().([]byte),
			Headers:   msg.FieldByName("Headers").Interface().(map[string]string),
		}
	}

	return messages, nil
}
