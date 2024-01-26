package tests

import (
	"context"
	"time"

	"github.com/gothunder/thunder/internal/events/outbox/ent/entInit"
	"github.com/gothunder/thunder/internal/events/outbox/ent/entInit/enttest"
	"github.com/gothunder/thunder/pkg/events/outbox/message"
	"github.com/gothunder/thunder/pkg/events/outbox/storer"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"entgo.io/ent/dialect"
	_ "github.com/mattn/go-sqlite3"
)

var _ = Describe("outbox.Storer", func() {
	var dbClient *entInit.Client
	BeforeEach(func() {
		dbClient = enttest.Open(GinkgoT(), dialect.SQLite, "file:ent?mode=memory&cache=shared&_fk=1")
	})

	AfterEach(func() {
		if dbClient != nil {
			dbClient.Close()
		}
	})

	Describe("NewOutboxStorer", func() {
		When("called with no Options", func() {
			It("should return a non-nil Storer and nil error", func() {
				// Act
				storer, err := storer.NewOutboxStorer()

				// Assert
				Expect(err).ShouldNot(HaveOccurred())
				Expect(storer).ShouldNot(BeNil())
			})
		})
		When("called with Options", func() {
			It("should return a non-nil Storer and nil error", func() {
				// Arrange
				testsCases := [][]storer.StorerOptions{
					{
						storer.WithLogging(),
					},
					{
						storer.WithTracing(),
					},
					{
						storer.WithMetrics(),
					},
					{
						storer.WithLogging(),
						storer.WithTracing(),
					},
					{
						storer.WithLogging(),
						storer.WithMetrics(),
					},
					{
						storer.WithTracing(),
						storer.WithMetrics(),
					},
					{
						storer.WithLogging(),
						storer.WithTracing(),
						storer.WithMetrics(),
					},
				}

				for _, opts := range testsCases {
					// Act
					storer, err := storer.NewOutboxStorer(opts...)

					// Assert
					Expect(err).ShouldNot(HaveOccurred())
					Expect(storer).ShouldNot(BeNil())
				}
			})
		})
	})

	Describe("Storer.Store", func() {
		When("called with valid messages", func() {
			It("should return a nil error", func() {
				// Arrange
				storer, err := storer.NewOutboxStorer(
					storer.WithLogging(),
					storer.WithTracing(),
					storer.WithMetrics(),
				)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(storer).ShouldNot(BeNil())

				messages := []message.Message{
					{
						Topic:   "test",
						Payload: []byte("test"),
						Headers: map[string]string{},
					},
					{
						Topic:   "test2",
						Payload: []byte("test2"),
					},
				}

				// Act
				err = storer.Store(context.Background(), dbClient.OutboxMessage, messages)

				// Assert
				Expect(err).ShouldNot(HaveOccurred())
			})
			It("should store those messages in the database", func() {
				// Arrange
				storer, err := storer.NewOutboxStorer(
					storer.WithLogging(),
					storer.WithTracing(),
					storer.WithMetrics(),
				)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(storer).ShouldNot(BeNil())

				messages := []message.Message{
					{
						Topic:   "test",
						Payload: []byte("test"),
						Headers: map[string]string{},
					},
					{
						Topic:   "test2",
						Payload: []byte("test2"),
					},
				}

				// Act
				err = storer.Store(context.TODO(), dbClient.OutboxMessage, messages)

				// Assert
				Expect(err).ShouldNot(HaveOccurred())
				count, _ := dbClient.OutboxMessage.Query().Count(context.Background())
				Expect(count).Should(Equal(2))
				all, _ := dbClient.OutboxMessage.Query().All(context.Background())
				for _, msg := range all {
					Expect(msg.Topic).ShouldNot(BeEmpty())
					Expect(msg.Payload).ShouldNot(BeEmpty())
					Expect(msg.ID).ShouldNot(BeEmpty())
					Expect(msg.CreatedAt).To(BeTemporally("~", time.Now(), time.Second))
					Expect(msg.DeliveredAt.IsZero()).To(BeTrue())
					if msg.Topic == "test" {
						Expect(msg.Payload).To(Equal([]byte("test")))
					} else {
						Expect(msg.Topic).To(Equal("test2"))
						Expect(msg.Payload).To(Equal([]byte("test2")))
					}

				}
			})
		})
		When("called with invalid messages", func() {
			It("should return an error", func() {
				// Arrange
				storer, err := storer.NewOutboxStorer(
					storer.WithLogging(),
					storer.WithTracing(),
					storer.WithMetrics(),
				)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(storer).ShouldNot(BeNil())

				messages := [][]message.Message{
					{
						{
							Topic:   "",
							Payload: []byte("test"),
							Headers: map[string]string{},
						},
					},
					{
						{
							Topic:   "test2",
							Payload: []byte(""),
						},
					},
					{},
				}

				for _, msgs := range messages {
					// Act
					err = storer.Store(context.Background(), dbClient.OutboxMessage, msgs)

					// Assert
					Expect(err).Should(HaveOccurred())
				}
			})
		})
		When("called with a invalid client", func() {
			It("should return error", func() {
				// Arrange
				storer, err := storer.NewOutboxStorer(
					storer.WithLogging(),
					storer.WithTracing(),
					storer.WithMetrics(),
				)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(storer).ShouldNot(BeNil())

				messages := []message.Message{
					{
						Topic:   "test",
						Payload: []byte("test"),
						Headers: map[string]string{},
					},
					{
						Topic:   "test2",
						Payload: []byte("test2"),
					},
				}

				// Act
				err1 := storer.Store(context.Background(), nil, messages)
				err2 := storer.Store(context.Background(), dbClient, messages)
				err3 := storer.Store(context.Background(), 1, messages)

				// Assert
				Expect(err1).Should(HaveOccurred())
				Expect(err2).Should(HaveOccurred())
				Expect(err3).Should(HaveOccurred())
			})
		})
	})

	Describe("Storer.WithTxClient", func() {
		When("called with a valid client", func() {
			It("should return a non-nil TransactionalStorer", func() {
				// Arrange
				storer, err := storer.NewOutboxStorer(
					storer.WithLogging(),
					storer.WithTracing(),
					storer.WithMetrics(),
				)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(storer).ShouldNot(BeNil())

				// Act
				txStorer, err := storer.WithTxClient(dbClient.OutboxMessage)

				// Assert
				Expect(txStorer).ShouldNot(BeNil())
				Expect(err).ShouldNot(HaveOccurred())
			})
		})
		When("called with an invalid client", func() {
			It("should return a nil TransactionalStorer", func() {
				// Arrange
				storer, err := storer.NewOutboxStorer(
					storer.WithLogging(),
					storer.WithTracing(),
					storer.WithMetrics(),
				)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(storer).ShouldNot(BeNil())

				// Act
				txStorer, err := storer.WithTxClient(nil)

				// Assert
				Expect(txStorer).Should(BeNil())
				Expect(err).Should(HaveOccurred())
			})
		})

		Describe("TransactionalStorer.Store", func() {
			When("called with valid messages", func() {
				It("should return a nil error", func() {
					// Arrange
					storer, err := storer.NewOutboxStorer(
						storer.WithLogging(),
						storer.WithTracing(),
						storer.WithMetrics(),
					)
					Expect(err).ShouldNot(HaveOccurred())
					Expect(storer).ShouldNot(BeNil())

					txStorer, err := storer.WithTxClient(dbClient.OutboxMessage)
					Expect(err).ShouldNot(HaveOccurred())
					Expect(txStorer).ShouldNot(BeNil())

					messages := []message.Message{
						{
							Topic:   "test",
							Payload: []byte("test"),
							Headers: map[string]string{},
						},
						{
							Topic:   "test2",
							Payload: []byte("test2"),
						},
					}

					// Act
					err = txStorer.Store(context.Background(), messages)

					// Assert
					Expect(err).ShouldNot(HaveOccurred())
				})
				It("should store those messages in the database", func() {
					// Arrange
					storer, err := storer.NewOutboxStorer(
						storer.WithLogging(),
						storer.WithTracing(),
						storer.WithMetrics(),
					)
					Expect(err).ShouldNot(HaveOccurred())
					Expect(storer).ShouldNot(BeNil())

					txStorer, err := storer.WithTxClient(dbClient.OutboxMessage)
					Expect(err).ShouldNot(HaveOccurred())
					Expect(txStorer).ShouldNot(BeNil())

					messages := []message.Message{
						{
							Topic:   "test",
							Payload: []byte("test"),
							Headers: map[string]string{},
						},
						{
							Topic:   "test2",
							Payload: []byte("test2"),
						},
					}

					// Act
					err = txStorer.Store(context.Background(), messages)

					// Assert
					Expect(err).ShouldNot(HaveOccurred())
					count, _ := dbClient.OutboxMessage.Query().Count(context.Background())
					Expect(count).Should(Equal(2))
					all, _ := dbClient.OutboxMessage.Query().All(context.Background())
					for _, msg := range all {
						Expect(msg.Topic).ShouldNot(BeEmpty())
						Expect(msg.Payload).ShouldNot(BeEmpty())
						Expect(msg.ID).ShouldNot(BeEmpty())
						Expect(msg.CreatedAt).To(BeTemporally("~", time.Now(), time.Second))
						Expect(msg.DeliveredAt.IsZero()).To(BeTrue())
						if msg.Topic == "test" {
							Expect(msg.Payload).To(Equal([]byte("test")))
						} else {
							Expect(msg.Topic).To(Equal("test2"))
							Expect(msg.Payload).To(Equal([]byte("test2")))
						}

					}
				})
			})

			When("called with invalid messages", func() {
				It("should return an error", func() {
					// Arrange
					storer, err := storer.NewOutboxStorer(
						storer.WithLogging(),
						storer.WithTracing(),
						storer.WithMetrics(),
					)
					Expect(err).ShouldNot(HaveOccurred())
					Expect(storer).ShouldNot(BeNil())

					txStorer, err := storer.WithTxClient(dbClient.OutboxMessage)
					Expect(err).ShouldNot(HaveOccurred())
					Expect(txStorer).ShouldNot(BeNil())

					messages := [][]message.Message{
						{
							{
								Topic:   "",
								Payload: []byte("test"),
								Headers: map[string]string{},
							},
						},
						{
							{
								Topic:   "test2",
								Payload: []byte(""),
							},
						},
						{},
					}

					for _, msgs := range messages {
						// Act
						err = txStorer.Store(context.Background(), msgs)

						// Assert
						Expect(err).Should(HaveOccurred())
					}
				})
			})
		})
	})
})
