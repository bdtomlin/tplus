package tplus

import (
	"fmt"
	"html/template"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/gofiber/template/utils"
)

const (
	tplusToken = "<!--tplusContent-->"
	tplusHead  = "tplusHead--"
	tplusTail  = "tplusTail--"
)

// Engine struct
type Engine struct {
	// delimiters
	left  string
	right string
	// views folder
	directory string
	// http.FileSystem supports embedded files
	fileSystem http.FileSystem
	// views extension
	extension string
	// layout variable name that incapsulates the template
	layout string
	// reload on each render
	reload bool
	// debug prints the parsed templates
	debug bool
	// lock for funcmap and templates
	mutex sync.RWMutex
	// template funcmap
	funcmap map[string]interface{}
	// templates
	Templates *template.Template
}

// New returns a HTML render engine for Fiber
func New(directory, extension string) *Engine {
	engine := &Engine{
		left:      "{{",
		right:     "}}",
		directory: directory,
		extension: extension,
		layout:    tplusToken,
		funcmap:   make(map[string]interface{}),
	}
	return engine
}

func NewFileSystem(fs http.FileSystem, extension string) *Engine {
	engine := &Engine{
		left:       "{{",
		right:      "}}",
		directory:  "/",
		fileSystem: fs,
		extension:  extension,
		layout:     tplusToken,
		funcmap:    make(map[string]interface{}),
	}
	return engine
}

// Delims sets the action delimiters to the specified strings, to be used in
// templates. An empty delimiter stands for the
// corresponding default: {{ or }}.
func (e *Engine) Delims(left, right string) *Engine {
	e.left, e.right = left, right
	return e
}

// AddFunc adds the function to the template's function map.
// It is legal to overwrite elements of the default actions
func (e *Engine) AddFunc(name string, fn interface{}) *Engine {
	e.mutex.Lock()
	e.funcmap[name] = fn
	e.mutex.Unlock()
	return e
}

// Reload if set to true the templates are reloading on each render,
// use it when you're in development and you don't want to restart
// the application when you edit a template file.
func (e *Engine) Reload(enabled bool) *Engine {
	e.reload = enabled
	return e
}

// Debug will print the parsed templates when Load is triggered.
func (e *Engine) Debug(enabled bool) *Engine {
	e.debug = enabled
	return e
}

// Load the templates to the engine.
func (e *Engine) Load() error {
	// race safe
	e.mutex.Lock()
	defer e.mutex.Unlock()

	e.Templates = template.New(e.directory)

	// Set template settings
	e.Templates.Delims(e.left, e.right)
	e.Templates.Funcs(e.funcmap)

	// Loop trough each directory and register template files
	walkFn := func(path string, info os.FileInfo, err error) error {
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
		name = strings.TrimSuffix(name, e.extension)
		// Read the file
		// #gosec G304
		buf, err := utils.ReadFile(path, e.fileSystem)
		if err != nil {
			return err
		}

		// err = e.ParseTplusComponents(name, string(buf))
		// if err != nil {
		// 	return err
		// }

		err = e.ParseTplusTokens(name, string(buf))
		if err != nil {
			return err
		}
		// Debugging
		// if e.debug {
		// 	fmt.Printf("views: parsed template: %s\n", name)
		// }
		//fmt.Printf("[Engine] Registered view: %s\n", name)
		return err
	}

	if e.fileSystem != nil {
		return utils.Walk(e.fileSystem, e.directory, walkFn)
	}
	return filepath.Walk(e.directory, walkFn)
}

// ParseTplusTokens takes a template string and breaks it up into sub templates
// to be used for wrapping the tplusContent tokens
func (e *Engine) ParseTplusTokens(name string, content string) error {
	// fmt.Println(content)
	tpls := strings.Split(content, e.layout)
	switch len(tpls) {
	case 1:
		_, err := e.Templates.New(name).Parse(tpls[0])
		if e.debug {
			fmt.Printf("views: parsed template: %s\n", name)
		}
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
		if e.debug {
			fmt.Printf("views: parsed template: %s\n", tplusHead+name)
			fmt.Printf("views: parsed template: %s\n", tplusTail+name)
		}
		return nil
	case 3:
		return fmt.Errorf("too many tplus tokens in %v", name)
	}
	return fmt.Errorf("something went wrong parsing template %v", name)
}

// Render will render the template by name as well as wrap any layouts around it in left to right order
func (e *Engine) Render(out io.Writer, template string, binding interface{}, layouts ...string) error {
	if e.reload {
		if err := e.Load(); err != nil {
			return err
		}
	}

	for i := len(layouts) - 1; i >= 0; i-- {
		if err := e.Templates.ExecuteTemplate(out, tplusHead+layouts[i], binding); err != nil {
			return err
		}
	}
	if err := e.Templates.ExecuteTemplate(out, template, binding); err != nil {
		return err
	}
	for _, l := range layouts {
		if err := e.Templates.ExecuteTemplate(out, tplusTail+l, binding); err != nil {
			return err
		}
	}
	return nil
}
