package bkjw

import (
	"golang.org/x/net/html"
	"strings"
)

func document(data []byte) *html.Node { n, _ := html.Parse(strings.NewReader(string(data))); return n }
func attr(n *html.Node, key string) string {
	for _, a := range n.Attr {
		if a.Key == key {
			return a.Val
		}
	}
	return ""
}
func text(n *html.Node) string {
	var b strings.Builder
	var visit func(*html.Node)
	visit = func(x *html.Node) {
		if x.Type == html.TextNode {
			b.WriteString(x.Data)
			b.WriteByte(' ')
		}
		for y := x.FirstChild; y != nil; y = y.NextSibling {
			visit(y)
		}
	}
	visit(n)
	return strings.Join(strings.Fields(b.String()), " ")
}
func nodes(root *html.Node, predicate func(*html.Node) bool) []*html.Node {
	out := []*html.Node{}
	var visit func(*html.Node)
	visit = func(n *html.Node) {
		if predicate(n) {
			out = append(out, n)
		}
		for y := n.FirstChild; y != nil; y = y.NextSibling {
			visit(y)
		}
	}
	visit(root)
	return out
}
func tag(root *html.Node, name string) []*html.Node {
	return nodes(root, func(n *html.Node) bool { return n.Type == html.ElementNode && n.Data == name })
}
func hasClass(n *html.Node, class string) bool {
	for _, s := range strings.Fields(attr(n, "class")) {
		if s == class {
			return true
		}
	}
	return false
}
func cells(row *html.Node) []*html.Node {
	out := []*html.Node{}
	for n := row.FirstChild; n != nil; n = n.NextSibling {
		if n.Type == html.ElementNode && (n.Data == "td" || n.Data == "th") {
			out = append(out, n)
		}
	}
	return out
}
func rawText(n *html.Node) string {
	var b strings.Builder
	var walk func(*html.Node)
	walk = func(n *html.Node) {
		if n.Type == html.TextNode {
			b.WriteString(n.Data)
		}
		for c := n.FirstChild; c != nil; c = c.NextSibling {
			walk(c)
		}
	}
	walk(n)
	return b.String()
}
