package common

import (
	"fmt"
	"testing"

	"github.com/russross/blackfriday/v2"
)

func TestMarkDown(t *testing.T) {
	markdownText := `
# 标题1{#xxfref}

这是一个示例的 Markdown 文本。

## 标题2

这是一些正文内容。

- 列表项1
- 列表项2
- 列表项3
		`
	// 使用 blackfriday 解析 Markdown
	html := blackfriday.Run([]byte(markdownText))

	// 将 HTML 转为纯文本
	text := StripHTML(string(html))

	fmt.Println(text)
}

func TestMarkdownNodes(t *testing.T) {
	markdownText := `
# Guide

## Heading 2

### Heading 3

### Heading 3

## Heading 1

## 图片

## 代码块

## 普通

## 语法高亮支持

### 演示 Ruby 代码高亮

### 演示 Rails View 高亮
`
	ast := blackfriday.New().Parse([]byte(markdownText))
	html := blackfriday.Run([]byte(markdownText), blackfriday.WithRenderer(blackfriday.NewHTMLRenderer(blackfriday.HTMLRendererParameters{Flags: blackfriday.TOC})), blackfriday.WithExtensions(blackfriday.CommonExtensions))

	// 提取目录信息
	nodes := extractNodes(ast)

	fmt.Printf("%s\n", SeriToString(nodes))
	fmt.Printf("%s\n", html)

}
