package main

import (
	"encoding/json"
	"html/template"
	"io/ioutil"
	"os"
	"path/filepath"

	"goa.design/model/mdl"
)

// DefaultTemplate is the template used to render and serve diagrams by
// default.
const DefaultTemplate = `<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="utf-8">
	<title>{{ .Title }}</title>
	<style>
		{{ .CSS }}
	</style>
</head>
<body>
	<div class="title">
		{{ .Title }}
	</div>
	<div>
		<div class="description">
			{{ .Description }}
		</div>
		<div class="version" style="text: align-right">
			{{ .Version }}
		</div>
	</div>
	<div id="diagram"></div>
	<script src="http://localhost:35729/livereload.js"></script>
	<script src="https://cdn.jsdelivr.net/npm/mermaid/dist/mermaid.min.js"></script>
	<script>
		var mermaidAPI = mermaid.mermaidAPI;
		mermaidAPI.initialize({
			securityLevel: 'loose',
			theme: 'neutral',
			startOnLoad:false{{ if .MermaidConfig }},
			...{{ .MermaidConfig }}{{ end }}
		});
		var element = document.getElementById("diagram");
		var insertSvg = function(svgCode, bindFunctions) {
			element.innerHTML = svgCode;
		};
		var src = ` + "`{{ .MermaidSource }}`;" + `
		var graph = mermaidAPI.render("mermaid", src, insertSvg);
	</script>
</body>
</html>
`

// DefaultCSS is the CSS used to render and serve diagrams by default.
const DefaultCSS = `
body {
	padding: 10px;
	font-family: Arial;
}

.title {
	font-size: 120%;
	font-weight: bold;
	padding-bottom: 1em;
}

.version {
	font-size: 80%;
}

.element {
	font-family: Arial;
}

.element-title {
	font-weight: bold;
	padding-bottom: 0.8em;
}

.element-description {
	font-size: 80%;
}

.relationship {
	font-family: Arial;
	font-size: 80%;
	background-color: white;
}

.relationship-label {
}

.relationship-technology {
}
`

// ViewData is the data structure used to render the HTML template for a
// given view.
type ViewData struct {
	// Title of view
	Title string
	// Description of view
	Description string
	// Version of design
	Version string
	// MermaidSource is the Mermaid diagram source code.
	MermaidSource template.JS
	// MermaidConfig is the Mermaid config JSON.
	MermaidConfig template.JS
	// CSS rendered inline
	CSS template.CSS
}

// indexTmpl is the default Go template used to render views.
var indexTmpl = template.Must(template.New("view").Parse(DefaultTemplate))

// loadViews generates the views for the given Go package, loads and returns the
// results indexed by view keys.
func loadViews(pkg, out string, debug bool) (map[string]*mdl.RenderedView, error) {
	if err := gen(pkg, out, debug); err != nil {
		return nil, err
	}
	views := make(map[string]*mdl.RenderedView)
	err := filepath.Walk(out, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		b, err := ioutil.ReadFile(path)
		if err != nil {
			return err
		}
		var view mdl.RenderedView
		if err := json.Unmarshal(b, &view); err != nil {
			return err
		}
		views[view.Key] = &view
		return nil
	})
	if err != nil {
		return nil, err
	}
	return views, nil
}

// render generates the views and renders static pages from the results.
func render(pkg, config, out string, debug bool) error {
	views, err := loadViews(pkg, out, debug)
	if err != nil {
		return err
	}
	for _, view := range views {
		f, err := os.Create(filepath.Join(out, view.Key+".html"))
		if err != nil {
			return err
		}
		data := &ViewData{
			Title:         view.Title,
			Description:   view.Description,
			Version:       view.Version,
			MermaidSource: template.JS(view.Mermaid),
			MermaidConfig: template.JS(config),
			CSS:           template.CSS(DefaultCSS),
		}
		if err := indexTmpl.Execute(f, data); err != nil {
			return err
		}
	}
	return nil
}