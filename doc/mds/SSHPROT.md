# SSH协议

### 协议层次

包括三层

* 传输层
  * 与https协议通信大致一样，通过非对称的密钥，交换成对称的会话密钥
  * 非对称的密钥一般是rsa的公钥和私钥，对于ssh就是主机的密钥
  * 在防止中间人攻击有一些不同，https是通过CA证书，有权威机构来处理，对应ssh一般采用公布主机的fingerprint来让用户确认  
* 认证层
  * 一般是公钥认证、密码认证、键盘输入认证
* 连接隧道
  * 一般是session隧道
