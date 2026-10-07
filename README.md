# badgecade-pretendo-go

The auth and secure servers BadgeCade runs for Nintendo Badge Arcade. They're Pretendo's servers with some changes so they work with our proxy and save storage.

Pretendo's code is AGPL-3.0, so here's our modified version.

- `authentication/` is from [nintendo-badge-arcade-authentication](https://github.com/PretendoNetwork/nintendo-badge-arcade-authentication) (commit `54d3aee`)
- `secure/` is from [nintendo-badge-arcade-secure](https://github.com/PretendoNetwork/nintendo-badge-arcade-secure) (commit `69fb9ea`)

What we changed is in [MODIFICATIONS.md](MODIFICATIONS.md).

## Building

```bash
docker build -t badgecade-auth ./authentication
docker build -t badgecade-secure ./secure
```

Settings are env vars, check `example.env` in each folder.

## License

AGPL-3.0, same as the original. All credit for the original code goes to Pretendo Network.
