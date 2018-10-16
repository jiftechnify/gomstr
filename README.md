# gomstr
Reading Note of the book "Go Programming Blueprints", Chapter5, using mastodon streaming API instead of Twitter stream.

## What is this?
This repository contains a set of small programs which aggregate toots on mstdn.jp public stream. Toots are treated as "votes".

## Structure
- `mstdnvotes`:Ssubscribe to mastodon update event stream and if a toot contains specified "option", send it as "vote" to NSQ broker.
- `counter`: Subscribe to votes flow into NSQ broker and update vote counts(recorded in MongoDB) periodically. 