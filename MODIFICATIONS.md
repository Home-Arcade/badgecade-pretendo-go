# Modifications from upstream

As AGPL-3.0 section 5(a) requires, this file lists the changes Badgecade made to the original Pretendo Network code (October 2026).
Line-ending (CRLF/LF) changes across many files are not listed.

## authentication/

- `compat_login.go`, `nex_token.go` (new): custom handlers for `Login`, `LoginEx` and `RequestTicket`, including the guest account and NEX token handling.
- `main.go`: registers those handlers.
- `utility.go`: if no stored NEX account password exists for a PID, it derives one with HMAC-SHA256 from `NEX_ACCOUNT_SECRET`, matching the password the Badgecade HTTPS proxy gives the console.
- `database.go`: a missing account returns nil and other database errors are logged, instead of panicking.

## secure/

- `file_store.go`, `nex/datastore/file_store.go` (new): a signed-URL HTTPS file store backed by local storage, replacing S3.
- `nex/datastore/prepare_get_object.go`, `prepare_post_object.go`, `prepare_update_object.go`, `post_meta_binary.go`, `get_persistence_info.go`, `nintendo-badge-arcade/get_meta_by_owner_id.go`: DataStore object transfers go through that file store.
- `init.go`, `main.go`, `example.env`, `Dockerfile`: initialization and configuration for the file store.
- `prudp/connect.go`: bounds-checked parsing of the CONNECT payload, plus ticket/request HMAC, session key length and PID validation, so malformed packets are rejected instead of crashing the server.
