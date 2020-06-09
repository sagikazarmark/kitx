# GraphQL

[GraphQL](https://graphql.org/) is a query language for the web. For server-side projects mainly serving client-side
web applications GraphQL can be a great choice.

Similarly to the [grpc transport](https://github.com/go-kit/kit/tree/master/transport/grpc), this transport
is also limited to synchronous calls. Go kit at its core uses [RPC](https://en.wikipedia.org/wiki/Remote_procedure_call) for communication,
so subscriptions are not supported.

Go has several bindings for GraphQL, but [gqlgen](https://gqlgen.com/) seems to be the most advanced.
