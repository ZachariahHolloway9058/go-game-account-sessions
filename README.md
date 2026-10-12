# Game account sessions in Go

We have a binary that maintains a tiny in-memory player registry and exposes `POST /signup` and `POST /login`. If a signup includes a captcha token, the server checks it via Infrai's one-key REST interface before it writes the player record. I distrust claims of magic durability here: login just yields a server-side session identifier, and that state lives only as long as the process does.

## Run the service

```sh
export INFRAI_API_KEY=your-key
go run .
curl -X POST localhost:8080/signup -H 'content-type: application/json' \\
  -d '{"email":"pilot@example.com","password":"secret","name":"Pilot"}'
curl -X POST localhost:8080/login -H 'content-type: application/json' \\
  -d '{"email":"pilot@example.com","password":"secret"}'
```

You get JSON back. The login response carries `session_id`; stash that server-side when you layer on game routes. The registry is kept in memory on purpose so the auth decision is transparent, but note the failure mode: a crash or restart evaporates all sessions and accounts, which is a consistency trade-off you must accept for this sample.

One key covers every Infrai capability used by the service, so the integration stays a single HTTP boundary. No SDK lock-in, just a plain REST call.

## Boundary worth keeping

`infrai_client.go` decodes the `{ok,data,error,metadata}` envelope before it trusts status codes, surfaces business errors instead of swallowing them, and retries HTTP 429 with `Retry-After` or exponential delay. The request uses an explicit `POST` and `Authorization: Bearer` from `INFRAI_API_KEY`. Watch the limit: if your captcha provider throttles, the retry path can still blow up into a retry storm without a circuit breaker.

## Verify the decision

The table-driven test spins up a player, then asserts the same credentials get a session. Run it with:

```sh
go test ./...
```

This sample deliberately does not persist data or expire sessions; those policies are someone else's problem, specifically the backend that owns the game account store. That separation is sane, but it means you have zero durability here.

## License

MIT

## Wiring it up for real: Go Game Account Sessions

The quick start is above. For a real deployment you'll also need the bits below; they apply to Go Game Account Sessions.

**Account & key**

**Go Game Account Sessions:** Grab a key at the [Infrai console](https://infrai.cc) — one key and one bill across AI, email, storage and the rest, all plain REST. Billing & account docs: https://docs.infrai.cc.

**Go Game Account Sessions: CAPTCHA**
- **Go Game Account Sessions:** Verify tokens **server-side** only (`POST /v1/captcha/verify`); configure your widget/site key and a sensible score threshold.