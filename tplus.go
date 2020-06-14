package tplus

import (
	"errors"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const (
	tplusToken = "<!--tplusContent-->"
	tplusHead  = "tplusHead--"
	tplusTail  = "tplusTail--"
)

// Engine struct
type Engine struct {
	directory string
	extension string

	Templates *template.Template
}

// New returns a HTML render engine for Fiber
func New(directory, extension string, funcmap ...map[string]interface{}) *Engine {
	engine := &Engine{
		directory: directory,
		extension: extension,
		Templates: template.New(directory),
	}
	if len(funcmap) > 0 {
		engine.Templates.Funcs(funcmap[0])
	}

	if err := engine.Parse(); err != nil {
		log.Fatalf("html.New(): %v", err)
	}
	return engine
}

// Parse parses the templates to the engine.
func (e *Engine) Parse() error {
	// Loop trough each directory and register template files
	err := filepath.Walk(e.directory, func(path string, info os.FileInfo, err error) error {
		// Return error if exist
		if err != nil {
			return err
		}
		// Skip file if it's a directory or has no file info
		if info == nil || info.IsDir() {
			return nil
		}
		// Get file extension of file
		ext := filepath.Ext(path)
		// Skip file if it does not equal the given template extension
		if ext != e.extension {
			return nil
		}
		// Get the relative file path
		// ./templates/html/index.tmpl -> index.tmpl
		rel, err := filepath.Rel(e.directory, path)
		if err != nil {
			return err
		}
		// Reverse slashes '\' -> '/' and
		// partials\footer.tmpl -> partials/footer.tmpl
		name := filepath.ToSlash(rel)
		// Remove ext from name 'index.tmpl' -> 'index'
		name = strings.Replace(name, e.extension, "", -1)
		// Read the file
		// #gosec G304
		buf, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		// Create new template associated with the current one
		// This enable use to invoke other templates {{ template .. }}
		// _, err = e.Templates.New(name).Parse(string(buf))

		err = e.ParseTplusTokens(name, string(buf))
		if err != nil {
			return err
		}
		// Debugging
		//fmt.Printf("[Engine] Registered view: %s\n", name)
		return err
	})
	return err
}

// ParseTplusTokens takes a template string and breaks it up into sub templates
// to be used for wrapping the tplusContent tokens
func (e *Engine) ParseTplusTokens(name string, content string) error {
	// fmt.Println(content)
	tpls := strings.Split(content, tplusToken)
	switch len(tpls) {
	case 1:
		_, err := e.Templates.New(name).Parse(tpls[0])
		return err
	case 2:
		_, err := e.Templates.New(tplusHead + name).Parse(tpls[0])
		if err != nil {
			return err
		}
		_, err = e.Templates.New(tplusTail + name).Parse(tpls[1])
		if err != nil {
			return err
		}
		return nil
	case 3:
		return fmt.Errorf("too many tplus tokens in %v", name)
	}
	return fmt.Errorf("something went wrong parsing template %v", name)
}

// Render will render the template by name as well as wrap any layouts around it in left to right order
func (e *Engine) Render(out io.Writer, name string, binding interface{}, layouts ...string) error {
	if len(layouts) == 0 {
		return e.Templates.ExecuteTemplate(out, name, binding)
	}
	for i := len(layouts) - 1; i >= 0; i-- {
		if err := e.Templates.ExecuteTemplate(out, tplusHead+layouts[i], binding); err != nil {
			return err
		}
	}
	if err := e.Templates.ExecuteTemplate(out, name, binding); err != nil {
		return err
	}
	for _, l := range layouts {
		if err := e.Templates.ExecuteTemplate(out, tplusTail+l, binding); err != nil {
			return err
		}
	}
	// more performant than fmt.Errorf
	return errors.New(strings.Join([]string{"error executing template "}, name))
}
