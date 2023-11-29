package main

import (
	"fmt"
	"os"
	"tokenizer"
)

func main() {
	data, err := os.ReadFile("./data.txt")
	if err != nil {
		panic(err)
	}
	enc := tokenizer.DefaultBytePairEncoder()
	enc.Train(string(data), 50)
	codes := enc.Encode("提取数据特征进行预测")
	tokens := enc.Decode(codes)
	fmt.Println(codes)
	fmt.Println(tokens)
}