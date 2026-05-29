# span

span = secure pan = secure drive

Encrypted WebDAV proxy. Encrypts file names (AES-CBC) and content (AES-CTR) before
storing on an upstream WebDAV server (Alist, AliYunDrive, etc.). Local WebDAV server
decrypts transparently on read.

See [AGENTS.md](AGENTS.md) for architecture and development guide.
See [doc/secure.md](doc/secure.md) for the encryption model.
