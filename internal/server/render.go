package server

import (
	"bytes"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"path"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/web"
)

// Pagina son los datos que recibe toda página: el título (lo usa el layout base) y los datos propios.
type Pagina struct {
	Titulo string
	Datos  any
}

// plantillas tiene cada página ya combinada con el layout base, indexada por nombre de archivo
// (por ejemplo "inicio.html"). Cada página se parsea por separado porque todas definen "contenido".
var plantillas = cargarPlantillas()

func cargarPlantillas() map[string]*template.Template {
	archivos, err := fs.Glob(web.Templates, "templates/paginas/*.html")
	if err != nil {
		panic(err)
	}
	ps := make(map[string]*template.Template, len(archivos))
	for _, a := range archivos {
		ps[path.Base(a)] = template.Must(template.ParseFS(web.Templates, "templates/base.html", a))
	}
	return ps
}

// render muestra una página dentro del layout base con el código HTTP indicado. Primero la arma en
// memoria para no enviar una página a medias si la plantilla falla.
func render(w http.ResponseWriter, codigo int, pagina string, datos Pagina) {
	t, ok := plantillas[pagina]
	if !ok {
		log.Printf("render: no existe la plantilla %q", pagina)
		http.Error(w, "error al mostrar la página", http.StatusInternalServerError)
		return
	}
	var buf bytes.Buffer
	if err := t.ExecuteTemplate(&buf, "base", datos); err != nil {
		log.Printf("render %s: %v", pagina, err)
		http.Error(w, "error al mostrar la página", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(codigo)
	_, _ = buf.WriteTo(w)
}
