# Badgecade Pretendo Go services

Modified copies of Pretendo Network's Nintendo Badge Arcade NEX servers, as run by the Badgecade server.
They're published here to meet the source-offer requirement of the GNU AGPL-3.0 (section 13).

| Directory | Upstream project | Upstream commit |
|---|---|---|
| `authentication/` | [PretendoNetwork/nintendo-badge-arcade-authentication](https://github.com/PretendoNetwork/nintendo-badge-arcade-authentication) | `54d3aee` |
| `secure/` | [PretendoNetwork/nintendo-badge-arcade-secure](https://github.com/PretendoNetwork/nintendo-badge-arcade-secure) | `69fb9ea` |

See [MODIFICATIONS.md](MODIFICATIONS.md) for what was changed.

## Building

Each directory is its own Go module with a `Dockerfile`:

```bash
docker build -t badgecade-auth ./authentication
docker build -t badgecade-secure ./secure
```

Configuration goes through environment variables; see each directory's `example.env`.

## License

GNU Affero General Public License v3.0. See [LICENSE](LICENSE) and the `LICENSE` file in each directory.
The original copyright belongs to Pretendo Network and its contributors; Badgecade's changes are released under the same license.
