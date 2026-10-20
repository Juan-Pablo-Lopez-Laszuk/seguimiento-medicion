package server

import (
	"bytes"
	"errors"
	"fmt"
	"html/template"
	"io/fs"
	"log"
	"net/http"
	"path"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/effort"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/web"
)

// Pagina son los datos que recibe toda página: el título (lo usa el layout base) y los datos propios.
type Pagina struct {
	Titulo string
	Datos  any
}

// funciones son las funciones que pueden usar las plantillas.
var funciones = template.FuncMap{
	"fechaParaInput": fechaParaInput,
	"dict":           dict,
	"horas":          effort.FormatearHoras,       // horas con coma decimal, por ejemplo "2,5" (HU-25)
	"fechaHora":      domain.FechaHoraEnArgentina, // fecha y hora de Argentina, por ejemplo "20/10/2026 18:30" (HU-16b)
}

// dict arma un mapa con pares clave, valor para pasarle varios datos a un componente, por ejemplo
// {{template "confirmar" (dict "ID" "confirmar-baja-1" "Titulo" "…")}} (TEC-07).
func dict(pares ...any) (map[string]any, error) {
	if len(pares)%2 != 0 {
		return nil, errors.New("dict necesita pares de clave y valor")
	}
	m := make(map[string]any, len(pares)/2)
	for i := 0; i < len(pares); i += 2 {
		clave, ok := pares[i].(string)
		if !ok {
			return nil, fmt.Errorf("la clave %v de dict tiene que ser texto", pares[i])
		}
		m[clave] = pares[i+1]
	}
	return m, nil
}

// plantillas tiene cada página ya combinada con el layout base y los componentes (TEC-07), indexada por nombre de archivo
// (por ejemplo "inicio.html"). Cada página se parsea por separado porque todas definen "contenido".
var plantillas = cargarPlantillas()

func cargarPlantillas() map[string]*template.Template {
	archivos, err := fs.Glob(web.Templates, "templates/paginas/*.html")
	if err != nil {
		panic(err)
	}
	ps := make(map[string]*template.Template, len(archivos))
	for _, a := range archivos {
		ps[path.Base(a)] = template.Must(template.New(path.Base(a)).Funcs(funciones).ParseFS(web.Templates, "templates/base.html", "templates/componentes.html", a))
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
