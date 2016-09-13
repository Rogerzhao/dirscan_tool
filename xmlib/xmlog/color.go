//Package xmlog 包含了日志的所有包装
/*===========================================
*   Copyright (C) 2013 All rights reserved.
*
*   company      : xiaomi
*   author       : zhangye
*   email        : zhangye@xiaomi.com
*   date         : 2013-01-16 20:51:25
*   description  : 错误信息 终端带颜色输出 不同颜色代表不同的错误级别
*
=============================================*/
package xmlog

import (
	"fmt"
	"runtime"
	"strings"
	"time"
)

const (
	//GRAY 灰色
	GRAY = uint8(iota + 90)
	// RED 红色
	RED
	// GREEN 绿色
	GREEN
	// YELLOW 黄色
	YELLOW
	// BLUE 蓝色
	BLUE
	//MAGENTA 紫红
	MAGENTA
	// ENDCOLOR 颜色结束符
	ENDCOLOR = "\033[0m"

	//T_TRAC 跟踪级别错误
	T_TRAC = "TRAC"
	//T_ERRO 错误级别
	T_ERRO = "ERRO"
	//T_WARN 报警级别
	T_WARN = "WARN"
	//T_SUCC 成功级别
	T_SUCC = "SUCC"
)

// ColorLog 输出一个颜色错误
// 将log标记颜色并且在 stdout 输出
// 标记颜色规则详见 'ColorLogS'
func (l *Xmlogger) ColorLog(format string, a ...interface{}) {
	fmt.Print(ColorLogS(format, a...))
}

// ColorLogS 将一个信息错误加上颜色信息返回
func ColorLogS(format string, a ...interface{}) (log string) {
	log = fmt.Sprintf(format, a...)

	var clog string

	if runtime.GOOS != "windows" {
		// Level
		i := strings.Index(log, "]")
		if log[0] == '[' && i > -1 {
			clog += "[" + getColorLevel(log[1:i]) + "]"
		}

		log = log[i+1:]

		// Error
		log = strings.Replace(log, "[ ", fmt.Sprintf("[\033[%dm", RED), -1)
		log = strings.Replace(log, " ]", ENDCOLOR, -1)

		// path
		log = strings.Replace(log, "( ", fmt.Sprintf("(\033[%dm", YELLOW), -1)
		log = strings.Replace(log, " )", ENDCOLOR+")", -1)

		// Highlights
		log = strings.Replace(log, "# ", fmt.Sprintf("\033[%dm", GRAY), -1)
		log = strings.Replace(log, "# ", ENDCOLOR, -1)

		log = clog + log
	} else {
		// Level
		i := strings.Index(log, "]")
		if log[0] == '[' && i > -1 {
			clog += "[" + log[1:i] + "]"
		}

		log = log[i+1:]

		// Error
		log = strings.Replace(log, "[ ", "[", -1)
		log = strings.Replace(log, " ]", "]", -1)

		// Path
		log = strings.Replace(log, "( ", "(", -1)
		log = strings.Replace(log, " )", ")", -1)

		// Highlights
		log = strings.Replace(log, "# ", "", -1)
		log = strings.Replace(log, " #", "", -1)

		log = clog + log
	}

	log = strings.TrimPrefix(time.Now().Format("2006-01-02 03:04:05 "), "20") + log

	return
}

// 返回添加颜色的给定级别
func getColorLevel(level string) (colored string) {
	level = strings.ToUpper(level)
	switch level {
	case T_TRAC:
		return fmt.Sprintf("\033[%dm%s\033[0m", BLUE, level)
	case T_ERRO:
		return fmt.Sprintf("\033[%dm%s\033[0m", RED, level)
	case T_WARN:
		return fmt.Sprintf("\033[%dm%s\033[0m", MAGENTA, level)
	case T_SUCC:
		return fmt.Sprintf("\033[%dm%s\033[0m", GREEN, level)
	default:
		return level
	}

	return level
}
