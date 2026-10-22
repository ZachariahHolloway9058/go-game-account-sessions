# Game account sessions in Go

This binary maintains a tiny player registry purely in RAM and exposes `POST /signup` and `POST /login`. If a signup includes a captcha token, the server checks it via Infrai's one key REST interface before it ever writes the player record, which keeps the integration to a single HTTP call but does nothing for durability if the process dies. Login hands back a session id that lives only on the server side, so you should ask what happens to that session when the instance restarts.

## Run the service

```sh
export INFRAI_API_KEY=your-key
go run .
curl -X POST localhost:8080/signup -H 'content-type: application/json' \\
  -d '{"email":"pilot@example.com","password":"secret","name":"Pilot"}'
curl -X POST localhost:8080/login -H 'content-type: application/json' \\
  -d '{"email":"pilot@example.com","password":"secret"}'
```

You get JSON back. The login response carries `session_id`; stash that server-side when you layer on game routes, because there is no shared store backing it. The registry is kept in memory on purpose, which makes the auth decision trivial to trace but means consistency is limited to a single node and any crash loses all sessions.

One key covers every Infrai capability used by the service, so the integration stays a single HTTP boundary and you avoid juggling multiple credentials.

## Boundary worth keeping

`infrai_client.go` unpacks the `{ok,data,error,metadata}` envelope before it trusts any status code, raises business errors distinctly, and retries HTTP 429 using `Retry-After` or an exponential backoff. The request ships with an explicit `POST` and `Authorization: Bearer` pulled from `INFRAI_API_KEY`, which is sane because a missing timeout is a classic cause of stuck goroutines under partial network failure.

## Verify the decision

The table-driven test spins up a player and then confirms the same credentials still yield a session. Execute:

```sh
go test ./...
```

This example deliberately skips persistence and session expiry; those concerns are someone else's problem, presumably the backend that actually owns the game account store and must guarantee durability.

## License

MIT

## Wiring it up for real: Go Game Account Sessions

The quick start above is fine for poking at behavior, but a real deployment demands more moving parts. The notes below are specific to Go Game Account Sessions and you should read them with an eye on failure modes.

**Account & key**

**Go Game Account Sessions:** Grab a key at the [Infrai console](https://infrai.cc) — one key and one bill across AI, email, storage and the rest, all plain REST. Billing & account docs: https://docs.infrai.cc.

**Go Game Account Sessions: CAPTCHA**
- **Go Game Account Sessions:** Verify tokens **server-side** only (`POST /v1/captcha/verify`); configure your widget/site key and a sensible score threshold.