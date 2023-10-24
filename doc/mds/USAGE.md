# 使用指南

## 前置要求
* 所有server需要统一的用户，比如每一个server包含一个root用户、一个web用户
* 所有ecs都需要安装、并启动opensshd
* 需要把跳板机的id_rsa.pub 同步到所有server对应用户的authorized_keys中，并保证从跳板机可以ssh到所有server上

_**这些工作应该是工程化，通常在装机初始化时完成。**_

## 编译
### 源码编译
* git clone https://github.com/hanzhihua/yajs.git
* make

## 配置文件
_系统启动时需要制定 变量confdir，它设置了配置目录，相关配置文件都在该目录下面_

配置文件模板请参考 ： https://github.com/hanzhihua/yajs/tree/main/conf_template

配置文件有下面固定约定：
* user的公钥约定为confdir/pubs/username_pub
* sshuser的私钥约定为confdir/pris/username_rsa
_系统启动时会做上述检查，如不通过则启动失败_

配置文件包括：
* config.yaml 主配置文件
* acl_model、acl_policy 权限文件
  * 采用表达式方式，处理user、sshuser、server三者关系，及user和menu的两者关系
  * 鉴权策略: 至少有一个匹配的策略规则allow，并且没有deny
* yajs_hk 系统私钥
  * ssh协议需要，可通过ssh-keygen生成
* pubs 用户的公钥目录 
  * 如 username: 张三，那么就需要confdir/pubs/张三.pub
  * 公钥文件为ssh-keygen生成公钥
* pris sshuser的私钥目录
    * 如 username: 张三，那么就需要confdir/pris/张三_rsa
    * 私钥文件为ssh-keygen生成私钥，并同步到所有server的具体用户的authorized_keys文件中
#### config.yaml文件
主配置文件，包括user、sshuser、服务列表三部分
```
users:
  - username: admin
  - username: 张三
  - username: 李四
sshusers:
  - username: root
  - username: web
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
跳板机主要的鉴权，是user、sshuser、server三者关系，系统采用casbin来解决
acl_model.conf 权限模型文件，不需要修改

鉴权策略: 至少有一个匹配的策略规则allow，并且没有deny

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
p, normal_group, *|root, deny
p, normal_group, menu:jumpserver, deny
p, normal_group, menu:batchCmd, deny

g, test1, normal_group
g, test2, normal_group
g, admin, admin_group

```
* 说明
  * 菜单：使用"menu:"为前缀
  * server: 使用"server:"为前缀
  * server+sshuser: 使用"|"为分隔


## 运维
### 运行
可以使用系统服务方式,在/etc/systemd/system中新增yajs.service
```
[Unit]
Description=yajs server daemon
Wants=network-online.target
After=network-online.target

[Service]
Type=simple
ExecStart=/.../yajs -c 你的configdir

[Install]
WantedBy=multi-user.target
```
启动：systemctl start yajs.service

停止: systemctl stop yajs.service

查看日志：journalctl -u yajs.service

### 新增、删除、修改 用户、sshuser、server
* 修改config.yaml
* 如果新增用户 
  * 使用ssh-keygen生成公私钥，把公钥内容copy到pubs/*_pub文件
  * 私钥给到用户，做登录使用
* 修改权限 acl_policy.csv
* reload: kill -usr1 pid 


### 查看审计日志
* 登录到跳板机
* 审计日志文件在 confdir/logs/，文件名名是 r_用户名_日期.log
* cat 审计日志文件
