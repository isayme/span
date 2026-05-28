# 目标
项目对外提供 webdav 服务，客户端上传的文件会加密后上传到上游的 webdav 服务，最终实现数据加密存储。类比 http://mega.io/。

# 需求说明
## 文件名不加密

## 文件内容
1. 每个文件的加密密钥不同；
2. 相同文件内容加密后的密文不同；
3. 加密后的文件长度可以反推原始文件长度变化，比如密文文件有个固定的size增长；
4. 文件加密可以分块、并行加密，不能串行加密；
5. 加密后的文件可以通过seek随机访问，不需要从头解密文件；
6. 防篡改：加密后的文件内容可以做完整性校验；
7. 使用 AES 或类似算法；
8. 用户修改密码后不影响存量文件内容的解密；

## 数据安全
1. 数据存储在 boltdb；
2. 数据库不存储用户明文密码；
3. 用户密码需要经过 github.com/nbutton23/zxcvbn-go 库强弱检测，拒绝弱密码；
4. 即使三方获得了数据库，也无法解密文件内容；

# 概念
## password
用户密码，也称 user password / master password。用于身份认证和生成encryptKey、authKey。

## salt
随机 16 字节，CSPRNG 真随机。用于生成 encrypt key 和 auth key，明文持久化存储，不会被修改。

## master key
随机 16 字节，CSPRNG 真随机。不会被修改。

## encrypted master key
加密后的 master key，持久化存储。用户修改密码后会产生新的 encrypt key,encrypted master key 会被更新。

Encrypted Master Key = AES-ECB( Derived Encryption Key , Master Key )

## encrypted key
也称 derived encryption key, 由 password 和 salt 生成。用于加密 master key。

```
// Iterations 固定是 100000，Length 是 32。
Derived Key = PBKDF2-HMAC-SHA-512( Password , Salt , Iterations , Length )
```

## auth key / authentication Key
也称 derived auth key / derived authentication key, 由 password 和 salt 生成。用于登录认证。

## hashed auth key
auth key 不会明文持久化，会经过一次 sha256。

Hashed Authentication Key = SHA-256( Derived Authentication Key )

## file key
用于加密文件内容。

## encrypted file key
加密后的 file key。明文 file key 不会持久化存储。

Encrypted File Key = AES-ECB(Master Key, File Key)

## file content
待加密的文件内容。

## encrypted file content
加密后的文件内容。

Encrypted File Content = Encrypted File Key + AES-CTR(File Key, File Content)

## ~~file name~~
待加密的文件名。

## ~~encrypted file name~~
加密后的 file name。明文 file name 不会持久化存储。

Encrypted File Name = Encrypted File Key + AES-CBC(File Key, File Name)

## 技术文章
mega.io 的实现原理可以参考 https://www.macheng.im/post/e2ee-of-mega/
mega.io 官方安全白皮书 https://mega.nz/SecurityWhitepaper.pdf
