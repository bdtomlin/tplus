package tplus

import (
	"bytes"
	"regexp"
	"strings"
	"testing"

	"github.com/gofiber/template/html"
)

func trim(str string) string {
	return strings.TrimSpace(regexp.MustCompile(`\s+`).ReplaceAllString(str, " "))
}

var engine *Engine

func init() {
	engine = New("./templates", ".html")
	engine.Load()
}

func TestPartials(t *testing.T) {
	var buf bytes.Buffer
	engine.Render(&buf, "index", map[string]interface{}{
		"Title": "Hello, World!",
	})
	expect := `<h2>Header</h2> <h1>Hello, World!</h1> <h2>Footer</h2>`
	result := trim(buf.String())
	if expect != result {
		t.Fatalf("Expected:\n%s\nResult:\n%s\n", expect, result)
	}
}

func TestLayout(t *testing.T) {
	var buf bytes.Buffer
	engine.Render(&buf, "index", nil, "layouts/main")
	expect := `<!DOCTYPE html> <html> <head> <title>Main</title> </head> <body> <h2>Header</h2> <h1></h1> <h2>Footer</h2> </body> </html>`
	result := trim(buf.String())
	if expect != result {
		t.Fatalf("Expected:\n%s\nResult:\n%s\n", expect, result)
	}
}
func TestNestedLayout(t *testing.T) {
	var buf bytes.Buffer
	engine.Render(&buf, "index", nil, "layouts/nested", "layouts/main")
	expect := `<!DOCTYPE html> <html> <head> <title>Main</title> </head> <body> <div id="nested"> <h2>Header</h2> <h1></h1> <h2>Footer</h2> </div> </body> </html>`
	result := trim(buf.String())
	if expect != result {
		t.Fatalf("Expected:\n%s\nResult:\n%s\n", expect, result)
	}
}

func TestComponent(t *testing.T) {
	t.Skip()
	var buf bytes.Buffer
	engine.Render(&buf, "index", nil, "components/widget")
	expect := `<!DOCTYPE html> <html> <head> <title>Main</title> </head> <body> <div id="nested"> <h2>Header</h2> <h1></h1> <h2>Footer</h2> </div> </body> </html>`
	result := trim(buf.String())
	if expect != result {
		t.Fatalf("Expected:\n%s\nResult:\n%s\n", expect, result)
	}
}

func TestSingle(t *testing.T) {
	var buf bytes.Buffer
	engine.Render(&buf, "errors/404", map[string]interface{}{
		"Title": "Hello, World!",
	})
	expect := `<h1>Hello, World!</h1>`
	result := trim(buf.String())
	if expect != result {
		t.Fatalf("Expected:\n%s\nResult:\n%s\n", expect, result)
	}
}

func BenchmarkRenderLayout(b *testing.B) {
	var buf bytes.Buffer
	for n := 0; n < b.N; n++ {
		engine.Render(&buf, "index", nil, "layouts/main")
	}
}

func BenchmarkRender(b *testing.B) {
	var buf bytes.Buffer
	for n := 0; n < b.N; n++ {
		engine.Render(&buf, "index", nil)
	}
}

func BenchmarkRenderNestedLayout(b *testing.B) {
	var buf bytes.Buffer
	for n := 0; n < b.N; n++ {
		engine.Render(&buf, "index", nil, "layouts/nested", "layouts/main")
	}
}

func BenchmarkFiberRenderLayout(b *testing.B) {
	fe := html.New("./fibertemplates", ".html")
	fe.Load()
	var buf bytes.Buffer
	for n := 0; n < b.N; n++ {
		fe.Render(&buf, "index", nil, "layouts/main")
	}
}

func BenchmarkFiberRender(b *testing.B) {
	fe := html.New("./fibertemplates", ".html")
	fe.Load()
	var buf bytes.Buffer
	for n := 0; n < b.N; n++ {
		fe.Render(&buf, "index", nil)
	}
}
