package common

import (
	"fmt"
	"io"
	"strings"

	"github.com/russross/blackfriday/v2"
)

// Node 表示目录中的一个节点
type Node struct {
	Text  string `json:"text"`
	Link  string `json:"link"`
	Level int64  `json:"level"`
	Nodes []Node `json:"nodes,omitempty"`
}

type myRenderer struct {
	blackfriday.Renderer
}

func (r *myRenderer) RenderHeader(w io.Writer, ast *blackfriday.Node) {
	fmt.Println("==============")
	// 在这里实现生成锚点链接的逻辑
	if ast.Type == blackfriday.Heading {
		id := strings.ToLower(strings.ReplaceAll(extractText(ast), " ", "-"))
		w.Write([]byte(fmt.Sprintf(`<h%d id="%s">`, ast.HeadingData.Level, id)))
	}
}

// extractNodes 从 Markdown AST 中提取目录信息
func extractNodes(node *blackfriday.Node) []Node {
	var nodes []Node

	for child := node.FirstChild; child != nil; child = child.Next {
		if child.Type == blackfriday.Heading {
			text := extractText(child)
			link := "#" + strings.ToLower(strings.ReplaceAll(text, " ", "-"))
			nodes = append(nodes, Node{Text: text, Link: link})
		}
	}

	return nodes
}

// extractText 从节点中提取文本
func extractText(node *blackfriday.Node) string {
	var result string

	for child := node.FirstChild; child != nil; child = child.Next {
		if child.Type == blackfriday.Text {
			result += string(child.Literal)
		}
	}

	return result
}
