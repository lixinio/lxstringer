## LxStringer

自动生成枚举的status code和status name

在编码中， 经常会用一些枚举来定一个状态或者其他离散的类型， 比如

``` go
type FrozenStatus int16 // 保证金冻结状态

const (
	FrozenStatusUnknown FrozenStatus = iota
	FrozenStatusFreezing
	FrozenStatusUnfreeze
)
```

为了支持通过`FrozenStatus`转换成 `code` 或 `name`， 需要再编写额外难以维护的代码， 如

``` go
var FrozenStatusCodes = map[FrozenStatus]string{
	FrozenStatusUnknown:  "unknown",
	FrozenStatusFreezing: "freezing",
	FrozenStatusUnfreeze: "unfreeze",
}

var FrozenStatusNames = map[FrozenStatus]string{
	FrozenStatusUnknown:  "未知",
	FrozenStatusFreezing: "冻结中",
	FrozenStatusUnfreeze: "已解冻",
}

func (s FrozenAmountStatus) Code() string {
	if code, ok := FrozenAmountStatusCodes[s]; ok {
		return code
	}
	return strconv.Itoa(int(s))
}

func (s FrozenAmountStatus) Name() string {
	if code, ok := FrozenAmountStatusNames[s]; ok {
		return code
	}
	return strconv.Itoa(int(s))
}
```

如果需要`code` 转 `FrozenStatus` 就更复杂了

参照官方的[stringer](https://pkg.go.dev/golang.org/x/tools/cmd/stringer)工具， 魔改代码支持此特性

> 源码见[GITHUB](https://github.com/golang/tools/blob/master/cmd/stringer/stringer.go)

## 使用

可参考example目录 （*对照单元测试理解*）

只需要提供基本的定义， 例如
``` go
type S11 int

const (
	S11_1 S11     = iota // "A A" aaa
	S11_2                // "FD SAF" bbb
	S11_3                // "F发 生" ccc
	S11_4                // D DD ddd
	S11_5 = S11_4        // E EE eee
)
```
``` bash
# -type 需要自动生成代码的枚举变量
# example/s3.go 源文件
$  lxstringer -type=S31,S32,S33 example/s3.go
```

+ code值， 通过注释（按 `空格` 间隔）第一个表示
+ name值， 通过注释（按 `空格` 间隔）第二个表示
+ 如果code值或者name值有空格， 可以用双引号， 例如`"code example" "name example"`
+ 注释
  + 超过两个的部分被忽略
  + 如果没有注释， `code/name`内容用`类型的字符串`代替， 例如`S11_1`
  + 如果只有一段注释， `name`内容用`类型的字符串`代替， 例如`S11_1`

生成的代码用法如下
``` go
func TestS11(t *testing.T) {
	require.Equal(t, S11_1.Code(), "A A")
	require.Equal(t, S11_2.Code(), "FD SAF")
	require.Equal(t, S11_3.Code(), "F发 生")
	require.Equal(t, S11_4.Code(), "D")
	require.Equal(t, S11_5.Code(), "D")

	require.Equal(t, S11_1.Name(), "aaa")
	require.Equal(t, S11_2.Name(), "bbb")
	require.Equal(t, S11_3.Name(), "ccc")
	require.Equal(t, S11_4.Name(), "DD")
	require.Equal(t, S11_5.Name(), "DD")

	require.Equal(t, CodeToS11("A A", S11_1), S11_1)
	require.Equal(t, CodeToS11("FD SAF", S11_1), S11_2)
	require.Equal(t, CodeToS11("F发 生", S11_1), S11_3)
	require.Equal(t, CodeToS11("D", S11_1), S11_4)
}
```

## 其他参数
+ -code Code函数的名称，默认`Code`
+ -name Name函数的名称，默认`Name`
+ -code2id Code转枚举函数的名称，默认`CodeTo$Type$` 例如`CodeToS11`
  + 如果`-code2id=-` 会跳过生成
+ -output 输出文件， 默认是当前目录的`$OriginFileName$_string.go`
+ -skipcode code和name都取第一列， 适用不关心code， 只关心name的情形