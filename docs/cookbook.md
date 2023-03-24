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
