# Abstract

Multi-thread + sub-thread multi-target brute-force crack, support SOCKS5 proxy protocol (no authentication)

Burst Mode:

​				Specify the username and password file  
​				
​				Specify the username and password directly
​				

​				Specify username and password file + Specify username and password directly
​				
​				Default threads: 50  

--------------------------------------------------------------------------------------------------------------------

多线程 + 子线程多目标暴力破解，支持SOCKS5代理协议(无认证)

暴破模式:

​				指定用户名、密码文件

​				直接指定用户名、密码

​				指定用户名、密码文件+直接指定用户名、密码

​				默认线程 ：50
​				
# compile(编译):

```
go build BruteSSH.go 
```

# Usage(使用方法):

```
./BruteSSH -h

Usage of ./BruteSSH:
  -P string
    	Directly specified passwords	直接指定密码条
  -U string
    	Directly specified usernames	直接指定字典用户名
  -d int
    	Detail level (0/1)						是否显示细节(0否,1是)
  -h string
    	Target addresses							设置SSH服务器
  -p string
    	File containing passwords			指定密码文件
  -proxy string
    	SOCKS5 proxy address					设置SOCKS5代理
  -t int
    	Threads per address (default 50)	设置线程，默认50
  -u string
    	File containing usernames			指定用户名文件
```



```
eg:
  ./BruteSSH -h 127.0.0.1 -u u.txt -p p.txt -d 1 -t 50 
  ./BruteSSH -h 127.0.0.1,192.168.1.1 -u u.txt -p p.txt -d 1 -t 50 -U 414a -P 123456 -proxy 127.0.0.1:7890

```

