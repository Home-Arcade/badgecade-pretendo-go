# Changes

Changes we made to Pretendo's code (October 2026). A lot of files also show up as changed because of line endings, those aren't listed.

## authentication

- Added `compat_login.go` and `nex_token.go`: our own Login, LoginEx and RequestTicket handlers. The 3DS logs in with its own PID but uses the password our proxy gives it, so auth checks the proxy's signed token to figure out which password to use.
- Console PIDs get linked to the BadgeCade account that first logs in with them (`consolebindings` in Mongo).
- `utility.go`: if there's no stored password for a PID, it makes the same one the proxy does (HMAC-SHA256 with `NEX_ACCOUNT_SECRET`).
- `database.go`: doesn't panic anymore when an account is missing.

## secure

- Saves go to our own signed HTTPS file store instead of S3 (`file_store.go`, `nex/datastore/file_store.go` and the DataStore prepare/post/update handlers).
- Saves are stored under the player's BadgeCade account instead of the console PID, so the save is the same in both Nimbus modes and on other consoles.
- `GetPersistenceInfo` treats a save that never finished uploading as no save.
- `prudp/connect.go`: checks the CONNECT packet properly (lengths, ticket, PID) so bad packets get dropped instead of crashing the server.
- Config changes in `init.go`, `main.go`, `example.env` and `Dockerfile`.
