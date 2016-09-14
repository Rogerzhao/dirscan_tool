# dirscan_tool
实现了一个扫描指定目录，生成目录中文件sha1值的小工具。支持过滤掉特定的目录，或者过滤特定的文件。
文件的输出格式

使用方法：
  准备工作，设置$GOBIN=$GOPATH/bin
  export GOBIN=$GOPATH/bin
  运行以下命令
  go get github.com/Rogerzhao/dirscan_tool
  
  安装成功，在$GOBIN目录下找到编译好的可执行文件dirscan_tool
  
  运行方法：
  ./bin/dirscan_tool -c conf/dirscan.conf
  
  配置文件说明：dirscan.conf
  [log_conf]
  logdir=log  #日志文件所在目录，请先建立好这个目录
  prefix=dir_scan #日志文件的前缀

  [path]
  walkPath=/tmp/api_tantan #需扫描的目录
  filterDir=ratelimit.[a-d]? #目录的过滤条件，支持*, ? [] !等通配符
  filterFile=*.go          # 文件名过滤条件
  concurrentNumber=10      # 并发计算sha1的文件数量
  resultFile=/tmp/sha1.out # 扫描结果存放的文件名
  
  关于测试，可以到src/github.com/Rogerzhao/dirscan_tool源码目录下运行 go test
  会输出相应的代码测试结果
  
