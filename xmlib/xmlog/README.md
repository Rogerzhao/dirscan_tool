####################2014-02-14######################
赵成兵 {{zhaochengbing@xiaomi.com}}
在xmlog里面加入了邮件报警，邮件发送使用米聊的接口,欢迎使用。
    http://mailsvc.api.b2c.srv/send
加入的这个功能对原有的xmlog没有任何影响，另外定义了一个channel WarnChan
xmlog.go的62行嵌入了一个条件判断，并将出错信息导入到WarnChan中。
    if WarnMode == "ON" && l.prefix == "ERROR" {
            WarnChan <- nerr
     }

如上的条件判断，只有当WarnMode打开，且程序记入的ERROR级别的日志（通常出现这类错误是需要排查的）才会处理。
功能实现代码在warnpolicy.go里面，其余全无变化
##功能描述
    设定报警间隔，如30分钟。
    首次发现一个错误，立即发送一封邮件。
    时间间隔内，持续出现同一错误，不发送邮件，只是将出错次数加1，时间间隔结束再发一封。
    那么30分钟内有错误最多收到两封邮件。且不会遗漏。

##如何使用
    导入xmlog包后，
    在 go xmlog.WatchError()前面加入下面几行代码
    xmlog.ProgramName = "phonecity_locator"  // 主程序的取的一个易辨别的名字。
    xmlog.WarnMode = "ON"   // 打开报警模式
    xmlog.WarnInterval = 30 // 报警的时间间隔，单位为分。建议设置为15-60
    xmlog.MailReciver = GMailReciver    //报警邮件的接收者
    go xmlog.WatchWarn()
##运行效果
    截取了一封邮件


    
您好，位于10.237.93.164上的程序phonecity_locator发生运行错误.
时间: 2014-02-14 15:19:24
错误信息: request.go:58 ERROR Error 1146: Table 'gopher.ph_phone_ere1a' doesn't exist
频率: 1分钟内报告这个错误16次
如果错误未排除，1分钟后会再次发送报警邮件. 
###################2014-02-14##############################
