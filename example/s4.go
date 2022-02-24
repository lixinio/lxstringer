package example

type S41 int

const (
	S41_1 S41 = iota + 100 // "A b C" "d E f" "G h I"
	S41_2                  // "中 华" "人 们"
	S41_3                  // "啊`啊" "i'm ok"
)
