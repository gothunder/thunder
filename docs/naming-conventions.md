# Naming conventions

## Commands

Action + Subject
: e.g. `SendBnplStatusUpdateInProcessPartnerEmail`

## Queries/Resolvers

Action + Subject + Filter
: e.g. `FindOneUsubscriber`

## Events

TODO

## Entities

TODO

## GraphQL queries and resolvers

Who + Action + Subject + `s` if plural
: e.g. `PartnerListInvoiceTotals`, `SuperAdminListBNPLOffers`

## Controllers

TODO

- Input:
Action + Subject + Input
: e.g. `ListRecurringPaymentsInput`

## Repository

TODO: Should it be `typeRepo` or `typerepo`?
