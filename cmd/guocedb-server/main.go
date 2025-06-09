package main

import (
	"context"
	"fmt"
	"github.com/turtacn/guocedb/compute/parser"
)

func main() {
	// 创建一个新的 GuocedbParser 实例
	p := parser.NewGuocedbParser()

	// 示例 SQL 查询
	query := "SELECT * FROM users"

	// 解析 SQL 查询
	stmt, err := p.Parse(context.Background(), query)
	if err != nil {
		fmt.Printf("解析查询时出错: %v\n", err)
		return
	}

	// 打印解析结果
	fmt.Printf("解析的 AST: %v\n", stmt)
}