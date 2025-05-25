# nogo

A lightweight rate limiter with an entity-attribute style rule engine that can be used to filter requests.

Includes implementations for sliding, static window logs and fixed token buckets. The backing state store provided by default is Redis.

Currently used at [MakeMyTrip](https://www.makemytrip.com) & [Goibibo](https://www.goibibo.com) for their email and WhatsApp channels targeting a user base of 10 million daily active users.

## Usage:
1. Create rules that target a specific type of entity and specify timing constraints on its attribute key and value pairs.
1. Instantiate the client by specifying the strategy to be used along with configuration for the backing store. The default in-memory store is generally not scalable if the state is complex.
1. Each `LimitRequest` passed is considered in its entirety. That is to say that if a single attribute's constraints are not met, the request fails as a whole.

## Features:
1. Extensible for different sorts of limiting strategies as well the underlying storage required to store the state.
1. Configuration for rate limits can be updated on the fly without downtime.
1. Rule engine can solve many common filter based use-cases such as audience targeting.

## Examples for rules:
1. Let's say we want to restrict tokens for our latest AI model based on the tier (or attribute) of a user. Additionally, some models are deprecated and we're in the process of removing them so we don't want to burden infra team with spikes. We can also define overlapping rules for different contexts. Here is the rule config mirroring some of these requirements:

```json
{
    "user_id": {
        "type": "ID",
        "attributes": [
            {
                "description": "free users get 1000 tokens/minute, 25000 tokens/day",
                "type": "user_tier",
                "value": "free",
                "rates": [
                    {
                        "duration": 60000000000,
                        "limit": 1000
                    },
                    {
                        "duration": 84600000000000,
                        "limit": 25000
                    }
                ]
            },
            {
                "description": "pro users get 10000 tokens/minute, no cap on daily usage",
                "type": "user_tier",
                "value": "pro",
                "rates": [
                    {
                        "duration": 60000000000,
                        "limit": 10000
                    }
                ]
            }
        ]
    },
    "model_name": {
        "type": "model",
        "attributes": [
            {
                "description": "v2.1 model does not get more than 250 requests a day",
                "type": "type",
                "value": "v2.1",
                "rates": [
                    {
                        "duration": 84600000000000,
                        "limit": 250
                    }
                ]
            }
        ]
    }
}
```