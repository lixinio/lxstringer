package example

type S21 int
type S22 int

const (
	S21_1 S21 = iota // "A A" aaa
	S21_2            // "FD SAF" bbb
	S21_3            // "F发 生" ccc
)

const (
	S22_1 S22 = iota + 100 // "A b C" "d E f" "G h I"
	S22_2                  // "中 华" "人 们"
	S22_3                  // "啊`啊" "i'm ok"
)
