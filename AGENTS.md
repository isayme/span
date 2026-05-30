# span — Agent Guide

## What it is
Encrypted WebDAV proxy (`secure pan`). Wraps an upstream WebDAV service (Alist/AliYunDrive/etc.), encrypting all file **names** (AES-CBC) and file **content** (AES-CTR) before storing upstream. Local WebDAV server decrypts transparently on read.

## References
Security design doc: `doc/secure.md` (Chinese). Consult it when modifying crypto code.

## Commands
```sh
make dev              # CONF_FILE_PATH=./config/dev.yaml go run main.go
make test             # go test ./...
go run main.go -p 8080 -l info -v   # standalone
```

- Config via `CONF_FILE_PATH` env var (YAML). `config/*` is gitignored; `config/dev.yaml` is example.
- Version set via `-ldflags "-X span/internal.Version=x"` at build time.

## Entrypoints & boundaries

| Path | Role |
|------|------|
| `main.go` → `cmd.Execute()` | CLI entry (cobra) |
| `cmd/root.go` | Wires everything: flags, config, password, master key init, BoltDB, WebDAV server |
| `internal/filesystem.go` | Implements both `afero.Fs` (via `FileSystem`) and `webdav.FileSystem` (via `webdavFS`). Delegates to underlying `afero.Fs` (backed by `github.com/isayme/afero-webdav`). All methods use `defer` for access logging. |
| `internal/utils/password.go` | PBKDF2 key derivation, masterKey encrypt/decrypt, fileKey ops |

## Encryption model (see also `doc/secure.md`)

```
password (user input)
  └─ PBKDF2(password, salt, 100000, sha512) = 64B
       ├─ encryptKey (first 32B)  → AES-ECB encrypt/decrypt masterKey
       └─ authKey    (last 32B)   → stored in BoltDB for password verification
```

- **Master key** (16B random) — encrypted with encryptKey via AES-ECB, stored as `encryptedMasterKey` in BoltDB
- **File name encryption**: AES-CBC(masterKey, iv=sha256(name)[:16]) + PKCS5 padding → base64url
  - Deterministic: same plaintext name always produces same ciphertext (enables rename/dedup by the upstream)
- **File key** (16B random per file) — encrypted with masterKey via AES-ECB, prepended to upstream file content
- **File content encryption**: AES-CTR(fileKey, iv=blockPos-based)
  - Different positions with same content produce different ciphertext
  - Encrypted content on upstream = 16B encryptedFileKey + encryptChunks...
  - Read on upstream: fileSize - 16 (strip the fileKey prefix)
- Password strength checked with zxcvbn; score < 4 is rejected

## span.db (BoltDB)

Single bucket `"span"` stores three keys:
- `salt` — 16B random, generated on first login
- `authKey` — PBKDF2-derived, used to verify password on subsequent logins
- `encryptedMasterKey` — masterKey encrypted with encryptKey

Flow: wrong password → authKey mismatch → reject. Correct password → derive encryptKey → decrypt encryptedMasterKey → get masterKey.

## Key quirks
- Most log/error messages are in **Chinese**
- `Must*` functions panic on error (e.g., `MustEncryptFileName`, `MustRandomMasterKey`)
- Filesystem layer now uses `github.com/isayme/afero-webdav` (WebDAV-backed `afero.Fs`) instead of manual encryption in `fs.go`/`file.go`. Old `fileinfo.go`, `readablefile.go`, `writeablefile.go` have been deleted.
- Three AES modes used: ECB for key encryption (no IV needed), CBC for filenames (deterministic IV), CTR for content (position-based IV)
- User-Agent header set to `span/version` on upstream WebDAV requests

## Package layout

```
internal/            # Core logic
  conf.go            # Config struct
  filesystem.go      # afero.Fs + webdav.FileSystem impl (delegates to afero-webdav)
  ua.go              # User-Agent header
  version.go         # Build version
  zxcvbn.go          # Password strength checker
  constants/         # Shared constants
  config/            # Embedded config assets
  utils/             # Utility packages
    aes-cbc.go       # AES-CBC for file names
    aes-ctr.go       # AES-CTR for file content
    aes-ecb.go       # AES-ECB for key encryption
    aes-gcm.go       # AES-GCM (available for future use)
    aes-padding.go   # PKCS5 padding
    base64.go        # URL-safe base64
    hash.go          # SHA-256 helpers
    iv.go            # IV derivation (block-position based)
    kv.go            # Key-value helpers
    password.go      # PBKDF2 key derivation, masterKey/fileKey ops
    rand.go          # Crypto-random helpers
e2e/                 # End-to-end test
cmd/                 # CLI (cobra)
```
