package example

type S11 int

const (
	S11_1 S11     = iota // "A A" aaa
	S11_2                // "FD SAF" bbb
	S11_3                // "F发 生" ccc
	S11_4                // D DD ddd
	S11_5 = S11_4        // E EE eee
)
