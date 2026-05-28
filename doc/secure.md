# 一些说明

password: 用户认证用密码

salt: 随机 16 字节
encryptKey: pbkdf2(password, salt) 生成 64 字节的前 32 字节
authKey: pbkdf2(password, salt) 生成 64 字节的后 32 字节

masterKey: 主密码, 随机 16 字节。
encryptedMasterKey: aes-ecb 加解密 key: encryptKey
加密后的主密码，存储在 db 中，即使 db 被他人获取，没有正确的 password 也无法获取到 masterKey

fileKey: 随机 16 字节

# span.db 中存储的内容

salt
authKey
encryptedMasterKey

如果用户提供错误的 password，则无法计算比对出正确的 authKey，认证失败；
如果用户提供正确的 password，则正常计算比对出 authKey，同时获得 encryptKey，随即可解密 encryptedMasterKey 获得 masterKey；

# 文件名称不加密

# 文件内容加密

fileKey = 随机 16 字节
iv: 与当前写入的文件内容 pos 相关（这样同一个文件，即使是不同位置相同的文件内容加密后结果也不同）
encryptedFileKey = aes-ecb(masterKey, fileKey)
encryptChunk = aes-ctr(fileKey, content, iv)
加密后的文件内容 = encryptedFileKey + encryptChunk ...

加密后特性

-   相同内容文件密文不同；（文件更安全，被封杀也可以避免多个副本都受影响）
-   文件加密的密钥加密存储，即使获得密文也无法解密；
-   不同 chunk 的密文加解密没有上下文依赖；

# 不同的加密方式

aes-ecb 加密密钥（加密不需要 iv）
aes-ctr 加密文件内容（需要 iv，可以并行加密）
aes-cbc 加密文件名（需要 iv，后一个分组加密依赖前一个分组加密结果）
