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
