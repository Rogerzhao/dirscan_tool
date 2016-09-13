/*===========================================
*   Copyright (C) 2013 All rights reserved.
*
*   company      : xiaomi
*   author       : zhangye
*   email        : zhangye@xiaomi.com
*   date         : 2014-01-16 10:16:25
*   description  : 错误信息彩色输出单元测试
*
=============================================*/
package xmlog

import (
	"testing"
)

func TestColored(t *testing.T) {
	var testString = "test"
	// 输出 14-01-16 10:26:21 [ERRO] Test ERRO test 颜色为红色
	DefaultLog.ColorLog("[ERRO] Test ERRO %s\n", testString)

	// 输出 14-01-16 10:26:21 [TRAC] Test TRAC test 颜色为蓝色
	DefaultLog.ColorLog("[TRAC] Test TRAC %s\n", testString)

	// 输出 14-01-16 10:26:21 [WARN] Test WARN test 颜色为品红色
	DefaultLog.ColorLog("[WARN] Test WARN %s\n", testString)

	// 输出 14-01-16 10:26:21 [SUCC] Test SUCC test 颜色为绿色
	DefaultLog.ColorLog("[SUCC] Test SUCC %s\n", testString)

	DefaultLog.ColorLog("[SUCC] ( path ) # SUCC %s\n", testString)
}
