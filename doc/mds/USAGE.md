# 使用指南

## 前置要求
* 所有ecs需要统一的用户，比如每一个ecs包含一个root用户、一个web用户
* 所有ecs都需要安装、并启动opensshd
* 需要把跳板机的id_rsa.pub 同步到所有ecs对应用户的authorized_keys中，并保证从跳板机可以ssh到所有ecs上

_**这些工作应该是工程化，通常在装机初始化时完成。**_

## 配置文件
_启动变量confdir是配置目录，相关配置文件都在该目录下面_
配置文件包括：
* config.yaml 主配置文件
* acl_model、acl_policy 权限文件
* yajs_hk 系统私钥（ssh协议需要，可通过ssh-keygen生成）
* pubs 用户的公钥目录 (如 username: 张三，那么就需要confdir/pubs/张三.pub)

#### config.yaml
主配置文件，包括yajs用户、统一的ssh user、服务列表三部分
```
users:
  - username: admin
  - username: 张三
  - username: 李四
sshusers:
  -
    username: root
    privateKeyFile: root_rsa
  -
    username: web
    privateKeyFile: web_rsa
servers:
  -
    name: saas01
    ip: 10.1.1.1
  -
    name: saas02
    ip: 10.1.1.2
    port: 22
    
```
其中servers也支持provider方式，目前提供了aliyun server provider，方式如下
```
server_provider:  aliyun
```

### 权限文件
跳板机主要的鉴权，是user、sshuser、server三者关系，yajs系统采用casbin来解决
acl_model.conf 权限模型文件，不需要修改
```
[request_definition]
r = user, resource

[policy_definition]
p = user, resource,eft

[role_definition]
g = _, _

[policy_effect]
e = some(where (p.eft == allow)) && !some(where (p.eft == deny))

[matchers]
m = g(r.user,p.user) && rMatch(r.resource, p.resource)
```

acl_policy.csv 权限策略文件
```
p, admin_group, *, allow
p, normal_group, *, allow
p, normal_group, menu:jumpserver, deny

g, 张三, normal_group
g, 李四, normal_group
g, admin, admin_group

```

### 运行
可以使用系统服务方式,新增 yajs.service
```
[Unit]
Description=yajs server daemon
Wants=network-online.target
After=network-online.target

[Service]
Type=simple
Environment="ACCESS_KEY_ID=你的KEY"
Environment="ACCESS_KEY_SECRET=你的KEY_SECRET"
ExecStart=/data/server/yajs.service/yajs -c /data/server/yajs.service

[Install]
WantedBy=multi-user.target
```
启动：systemctl start yajs.service

停止: systemctl stop yajs.service

查看日志：journalctl -u yajs.service


### 查看审计日志
* 登录到跳板机
* 审计日志文件在 confdir/logs/，文件名名是 r_用户名_日期.log
* cat 审计日志文件
