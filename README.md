# yajs
Yet Another Jump Server
![img.png](doc/imgs/overview.png)

## 动机
生产环境中有几十台ecs服务器，很明显是需要一个跳板机来支撑日常运维工作，在github.com找了一些开源解决方案，并没特别合适的,
都需要花费一定时间做二次开发。

目前我的时间有点充裕，跳板机系统涉及到的技术，包括ssh协议、pty、shell，我也比较有兴趣，而且我对网络开发、数据安全有一定的经验，所以决定自己重新做一个。

## 系统功能
* yajs实现了跳板机所有基本功能，目前已在线上运行了几个月
* 主界面采用了menu ui方式，功能已满足需求  
* 运维安全
  * 鉴权和授权：包括user、server、sshuser三者、及user和menu二者，系统使用rbac+鉴权表达式方式来处理  
  * 认证：只实现了public key方式 
  * 审计：用户的操作日志已记录  


![img.png](doc/imgs/feature.png)

## 使用指南
使用指南相关的，请移步到[使用指南](doc/mds/USAGE.md).


## 系统设计
系统设计相关的，请移步到[系统设计](doc/mds/DESIGN.md).

## 遇到到问题
系统开发中遇到到问题，及解决方案也记录下来，有兴趣请移步到[遇到到问题](doc/mds/FAQ.md).