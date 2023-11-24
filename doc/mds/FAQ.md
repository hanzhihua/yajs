# 遇到的问题

### 进入具体server时，shell丢失相关的环境变量
解决方案：
    * 原来代码是这样写的 
  ```
  shell := os.Getenv("SHELL")
	if shell == ""{
		shell = "/bin/bash"
	}
	cmd := exec.Command(shell)
  ```
  最后一行改成
  
  ```
  cmd := exec.Command(shell,"--login")
  ```

  原来的shell是交互不登录的Shell,相关的profile，没有执行，改成了login方式，profile会被执行。 

### 从服务器退出后，进入跳板机主界面，第一次按键不生效问题解决

* 分析：
    * 跳板机主界面的菜单方式，输入、输出都是ssh session
    * 当进入服务器后，输入、输出也同样是ssh session
    * 从服务器退出时，输入并没有直接关闭，因为退出是block操作，只有等下一次按键触发eof错误，关闭输入，被主主界面接管
        * 这个时候如果有线程中断该多好

* 解决方案：
    * 增强session功能，当reader有多个数据接收方时，保留最后一次输入的读取，
    * 后来的数据接收方，从保留最后输入中读取数据

