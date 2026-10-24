package features

import (
	"bytes"
	"compress/zlib"
	"context"
	"fmt"
	"html"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"time"

	"github.com/cucumber/godog"
	"golang.org/x/text/encoding/charmap"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/sprint"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/reportepdf"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/server"
)

// hu39 guarda el estado de un escenario de HU-39 (exportar el reporte a PDF). Usa los repositorios y los casos de uso
// de HU-37 y HU-38; el PDF y la página del mismo reporte se piden a la aplicación, como los botones.
type hu39 struct {
	*hu38
	generado  time.Time                  // cuándo se genera el PDF
	descarga  *httptest.ResponseRecorder // la respuesta de la descarga
	demora    time.Duration              // cuánto tardó la descarga
	documento reportepdf.Documento       // lo que la aplicación dibuja en el PDF
	pantalla  string                     // la página del mismo reporte
}

// registrarPasosHU39 conecta cada frase de features/hu-39-exportar-pdf.feature con una función de Go.
func registrarPasosHU39(sc *godog.ScenarioContext) {
	e := &hu39{hu38: &hu38{hu37: nuevoHU37(), proyectos: map[string]int64{}}}

	sc.Step(`^el proyecto "([^"]*)" con estos sprints para el PDF:$`, e.elProyectoConEstosSprints)
	sc.Step(`^estas historias del proyecto para el PDF:$`, e.estasHistoriasDelProyecto)
	sc.Step(`^que para el PDF el Sprint (\d+) se cerró con "([^"]*)" y completó "([^"]*)"$`, e.queElSprintSeCerroConYCompleto)
	sc.Step(`^que para el PDF "([^"]*)" está en el Sprint (\d+)$`, e.queEstaEnElSprint)
	sc.Step(`^que el PDF se genera el (\d+)/(\d+)/(\d+)$`, e.queElPDFSeGeneraEl)
	sc.Step(`^el proyecto "([^"]*)" con (\d+) sprints cerrados de (\d+) historias cada uno$`, e.elProyectoGrande)
	sc.Step(`^que "([^"]*)" se llama "([^"]*)"$`, e.queSeLlama)
	sc.Step(`^descargo el PDF del reporte del Sprint (\d+)$`, e.descargoElPDFDelSprint)
	sc.Step(`^descargo el PDF del reporte del proyecto$`, e.descargoElPDFDelProyecto)
	sc.Step(`^descargo el PDF del reporte del proyecto "([^"]*)"$`, e.descargoElPDFDelProyectoLlamado)
	sc.Step(`^se descarga el archivo "([^"]*)"$`, e.seDescargaElArchivo)
	sc.Step(`^el PDF muestra lo mismo que el reporte del (?:Sprint \d+|proyecto) en pantalla$`, e.elPDFMuestraLoMismoQueLaPantalla)
	sc.Step(`^el PDF dice "([^"]*)"$`, e.elPDFDice)
	sc.Step(`^el PDF muestra la historia "([^"]*)"$`, e.elPDFDice)
	sc.Step(`^cada sprint del PDF empieza en una hoja nueva$`, e.cadaSprintEmpiezaEnUnaHojaNueva)
	sc.Step(`^el PDF se genera en menos de (\d+) segundos?$`, e.elPDFSeGeneraEnMenosDe)
	sc.Step(`^la descarga responde (\d+) con "([^"]*)"$`, e.laDescargaResponde)
}

func (e *hu39) queElPDFSeGeneraEl(dia, mes, anio int) error {
	e.generado = time.Date(anio, time.Month(mes), dia, 18, 0, 0, 0, time.UTC) // 15:00 en la Argentina
	return nil
}

// elProyectoGrande arma un proyecto con sprints semanales, cada uno con sus historias (la mitad Hechas) y cerrado.
func (e *hu39) elProyectoGrande(nombre string, sprints, historias int) error {
	ctx := context.Background()
	inicio := fechaDeEscenario("2026-01-05")
	p, err := e.repoProyectos.Guardar(ctx, project.Proyecto{Nombre: nombre, FechaInicio: inicio, FechaFin: fechaDeEscenario("2026-12-31")})
	if err != nil {
		return err
	}
	e.proyectoID, e.proyectos[nombre] = p.ID, p.ID
	for n := 1; n <= sprints; n++ {
		s, err := e.repoSprints.Guardar(ctx, sprint.Sprint{
			ProyectoID: p.ID, Numero: n, Objetivo: fmt.Sprintf("Objetivo del Sprint %d", n),
			FechaInicio: inicio.AddDate(0, 0, 7*(n-1)), FechaFin: inicio.AddDate(0, 0, 7*(n-1)+6), Estado: sprint.EstadoPlanificado,
		})
		if err != nil {
			return err
		}
		e.sprintIDs[n], e.fines[n] = s.ID, s.FechaFin
		for i := 1; i <= historias; i++ {
			sp, numero := 3, (n-1)*historias+i
			h, err := e.repoHistorias.Guardar(ctx, backlog.Historia{
				ProyectoID: p.ID, Numero: numero, Prioridad: backlog.PrioridadAlta, Estado: backlog.EstadoPendiente, StoryPoints: &sp,
				Titulo: fmt.Sprintf("Historia %d con un título largo para que ocupe más de una línea en la tabla del PDF", numero),
			})
			if err != nil {
				return err
			}
			estado := backlog.EstadoEnSprint
			if i%2 == 0 {
				estado = backlog.EstadoHecho
			}
			if err := e.repoHistorias.ActualizarSprint(ctx, h.ID, estado, &s.ID); err != nil {
				return err
			}
		}
		if err := e.queElSprintSeCerro(n); err != nil {
			return err
		}
	}
	return nil
}

func (e *hu39) queSeLlama(id, titulo string) error {
	n, err := numeroDe(id)
	if err != nil {
		return err
	}
	ctx := context.Background()
	h, err := e.repoHistorias.BuscarPorNumero(ctx, e.proyectoID, n)
	if err != nil {
		return err
	}
	h.Titulo = titulo
	return e.repoHistorias.ActualizarDatos(ctx, h)
}

func (e *hu39) descargoElPDFDelSprint(numero int) error {
	ruta := fmt.Sprintf("/proyectos/%d/sprints/%d/reporte", e.proyectoID, numero)
	if !e.descargar(ruta) {
		return nil // el escenario de error mira la respuesta
	}
	reporte, err := e.reportes.Sprint(context.Background(), e.proyectoID, numero)
	e.documento = reportepdf.DocumentoSprint(reporte)
	return err
}

func (e *hu39) descargoElPDFDelProyecto() error {
	return e.descargarProyecto(e.proyectoID)
}

func (e *hu39) descargoElPDFDelProyectoLlamado(nombre string) error {
	id, ok := e.proyectos[nombre]
	if !ok {
		return fmt.Errorf("no está el proyecto %q", nombre)
	}
	return e.descargarProyecto(id)
}

func (e *hu39) descargarProyecto(proyectoID int64) error {
	if !e.descargar(fmt.Sprintf("/proyectos/%d/reporte", proyectoID)) {
		return nil
	}
	reporte, err := e.reportes.Proyecto(context.Background(), proyectoID)
	e.documento = reportepdf.DocumentoProyecto(reporte)
	return err
}

// descargar pide el PDF del reporte (la ruta con .pdf) y la página del mismo reporte. Devuelve si la descarga dio 200.
func (e *hu39) descargar(reporte string) bool {
	h := server.NewRouter(server.Dependencias{Reportes: e.reportes, Ahora: func() time.Time { return e.generado }})
	e.descarga = httptest.NewRecorder()
	inicio := time.Now()
	h.ServeHTTP(e.descarga, httptest.NewRequest(http.MethodGet, reporte+".pdf", nil))
	e.demora = time.Since(inicio)
	pagina := httptest.NewRecorder()
	h.ServeHTTP(pagina, httptest.NewRequest(http.MethodGet, reporte, nil))
	e.pantalla = pagina.Body.String()
	return e.descarga.Code == http.StatusOK
}

func (e *hu39) seDescargaElArchivo(archivo string) error {
	if e.descarga.Code != http.StatusOK {
		return fmt.Errorf("la descarga respondió %d: %s", e.descarga.Code, e.descarga.Body.String())
	}
	if ct := e.descarga.Header().Get("Content-Type"); ct != "application/pdf" {
		return fmt.Errorf("Content-Type = %q, se esperaba application/pdf", ct)
	}
	if cd, esperado := e.descarga.Header().Get("Content-Disposition"), `attachment; filename="`+archivo+`"`; cd != esperado {
		return fmt.Errorf("Content-Disposition = %q, se esperaba %q", cd, esperado)
	}
	return nil
}

// soloEnElPDF son los textos que no están en la página: los encabezados de las tablas de métricas, que en la página
// son tarjetas.
var soloEnElPDF = map[string]bool{"Métrica": true, "Valor": true}

var espacios = regexp.MustCompile(`\s+`)

func juntar(textos []string) string {
	return espacios.ReplaceAllString(strings.Join(textos, " "), " ")
}

// elPDFMuestraLoMismoQueLaPantalla compara en las dos direcciones los textos de la página (sin la navegación ni los
// botones) con lo que la aplicación dibuja en el PDF, y revisa que el PDF descargado tenga cada uno (CA-39.2).
func (e *hu39) elPDFMuestraLoMismoQueLaPantalla() error {
	nodos := textosDeLaPagina(e.pantalla)
	documento, pagina, pdf := juntar(e.documento.Textos()), juntar(nodos), textoDelPDF(e.descarga.Body.Bytes())
	for _, nodo := range nodos {
		if !strings.Contains(documento, nodo) {
			return fmt.Errorf("la página muestra %q y el PDF no", nodo)
		}
	}
	for _, texto := range e.documento.Textos() {
		texto = juntar([]string{texto})
		if !soloEnElPDF[texto] && !strings.Contains(pagina, texto) {
			return fmt.Errorf("el PDF tiene %q y la página no", texto)
		}
		if !strings.Contains(pdf, texto) {
			return fmt.Errorf("el PDF descargado no tiene %q", texto)
		}
	}
	return nil
}

func (e *hu39) elPDFDice(texto string) error {
	if !strings.Contains(textoDelPDF(e.descarga.Body.Bytes()), texto) {
		return fmt.Errorf("el PDF descargado no dice %q", texto)
	}
	return nil
}

// cadaSprintEmpiezaEnUnaHojaNueva busca el título de cada sprint como primer texto de alguna hoja del PDF descargado.
func (e *hu39) cadaSprintEmpiezaEnUnaHojaNueva() error {
	primeros := map[string]bool{}
	for _, hoja := range hojasDelPDF(e.descarga.Body.Bytes()) {
		if textos := textosDeLaHoja(hoja); len(textos) > 0 {
			primeros[textos[0]] = true
		}
	}
	for _, b := range e.documento.Bloques {
		if b.Tipo == reportepdf.Subtitulo && !primeros[b.Texto] {
			return fmt.Errorf("%q no empieza una hoja nueva", b.Texto)
		}
	}
	return nil
}

func (e *hu39) elPDFSeGeneraEnMenosDe(segundos int) error {
	if e.descarga.Code != http.StatusOK {
		return fmt.Errorf("la descarga respondió %d", e.descarga.Code)
	}
	if limite := time.Duration(segundos) * time.Second; e.demora >= limite {
		return fmt.Errorf("el PDF tardó %v, se esperaba menos de %v", e.demora, limite)
	}
	return nil
}

func (e *hu39) laDescargaResponde(codigo int, mensaje string) error {
	if e.descarga.Code != codigo || !strings.Contains(e.descarga.Body.String(), mensaje) {
		return fmt.Errorf("se esperaba %d con %q y se obtuvo %d con %q", codigo, mensaje, e.descarga.Code, e.descarga.Body.String())
	}
	return nil
}

var (
	quitarDeLaPagina = regexp.MustCompile(`(?s)<nav.*?</nav>|<button.*?</button>|<a [^>]*d-print-none[^>]*>.*?</a>`)
	etiqueta         = regexp.MustCompile(`<[^>]*>`)
)

// textosDeLaPagina devuelve los textos de <main> sin la navegación ni los botones, que el PDF no lleva.
func textosDeLaPagina(pagina string) []string {
	main := pagina[strings.Index(pagina, "<main"):strings.Index(pagina, "</main>")]
	main = etiqueta.ReplaceAllString(quitarDeLaPagina.ReplaceAllString(main, ""), "\n")
	var textos []string
	for _, linea := range strings.Split(html.UnescapeString(main), "\n") {
		if linea = strings.TrimSpace(espacios.ReplaceAllString(linea, " ")); linea != "" {
			textos = append(textos, linea)
		}
	}
	return textos
}

var flujo = regexp.MustCompile(`(?s)stream\n(.*?)\nendstream`)

// hojasDelPDF devuelve el contenido de cada hoja del PDF descomprimido. fpdf solo comprime las hojas; las fuentes base
// no van en el archivo.
func hojasDelPDF(pdf []byte) []string {
	var hojas []string
	for _, m := range flujo.FindAllSubmatch(pdf, -1) {
		r, err := zlib.NewReader(bytes.NewReader(m[1]))
		if err != nil {
			continue
		}
		contenido, err := io.ReadAll(r)
		if err == nil {
			hojas = append(hojas, string(contenido))
		}
	}
	return hojas
}

var textoEscrito = regexp.MustCompile(`\(((?:\\.|[^\\)])*)\)\s*Tj`)

// textosDeLaHoja devuelve, en orden, los textos que escribe una hoja, pasados de Windows-1252 a UTF-8.
func textosDeLaHoja(hoja string) []string {
	var textos []string
	for _, m := range textoEscrito.FindAllStringSubmatch(hoja, -1) {
		crudo := strings.NewReplacer(`\\`, `\`, `\(`, "(", `\)`, ")").Replace(m[1])
		texto, err := charmap.Windows1252.NewDecoder().String(crudo)
		if err == nil {
			textos = append(textos, texto)
		}
	}
	return textos
}

// textoDelPDF junta todos los textos del PDF descargado. Un texto partido en líneas queda unido con espacios.
func textoDelPDF(pdf []byte) string {
	var textos []string
	for _, hoja := range hojasDelPDF(pdf) {
		textos = append(textos, textosDeLaHoja(hoja)...)
	}
	return juntar(textos)
}
