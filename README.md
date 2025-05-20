# nogo

A lightweight rate limiter with an entity-attribute style rule engine that can be used to filter requests.

Includes implementations for sliding, static window logs and fixed token buckets. The backing state store provided by default is Redis.

## Usage:
1. Create rules that target a specific type of entity and specify timing constraints on its attribute key and value pairs.
1. Instantiate the client by specifying the strategy to be used along with configuration for the backing store.
1. Each `LimitRequest` passed is considered in its entirety. That is to say that if a single attribute's constraints are not met, the request fails as a whole.