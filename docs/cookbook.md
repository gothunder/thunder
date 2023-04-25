# Cookbook

## GraphQL

### Resolvers

- Add schema to `/internal/transport-inbound/graphql/<name>.graphql`. Note
  that the name of the file is arbitrary and impacts the name of the resolver
  file.
- Run `make generate` to generate the resolver file in
  `/internal/transport-inbound/graphql/resolvers/<name>.resolvers.go`.
- Write the code for the resolver in the generated file. Usually, this calls
  `commands` and `queries` modules in the `/internal/features`.

### Loaders

- Define the loader in `/internal/features/graphql/loaders/`.
- (Optional) Add the loader to the `Loaders` struct in
  `/internal/features/graphql/loaders/loaders.go`.
- Use the loader in `/internal/features/graphql/resolvers/entity.resolvers.go`.

## Repository Pattern

### ent

TODO

## GraphQL+ent integration

TODO

## Tests

### Mock subscribers

```go
subscriberCalled := make(chan interface{})
topic := events.SubscriberTopic
var subscriberReturn events.SubscriberMessage
handler.Mock.On("Handle",
    mock.Anything,
    topic,
    mock.Anything,
).Once().Return(thunderEvents.Success).Run(func(args mock.Arguments) {
    defer GinkgoRecover()

    decoder := args.Get(2).(thunderEvents.EventDecoder)
    err := decoder.Decode(&subscriberReturn)
    Expect(err).ToNot(HaveOccurred())

    close(subscriberCalled)
})

resp := someCommand() // this publishes an event

// Assert
Eventually(subscriberCalled, "3s").Should(BeClosed())
Expect(resp).To(Equal(thunderEvents.Success))
```
