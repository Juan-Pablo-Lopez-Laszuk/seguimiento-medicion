package reportepdf

import (
	"bytes"
	"fmt"
	"io"
	"strings"
	"time"

	"codeberg.org/go-pdf/fpdf"
	"golang.org/x/text/encoding/charmap"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain"
)

// Medidas de la hoja A4, en milímetros.
const (
	margen      = 15.0
	alturaLinea = 5.0
	relleno     = 1.0  // arriba y abajo del texto de cada celda
	alturaPie   = 12.0 // lo que ocupa el pie desde el borde de abajo
)

// Escribir dibuja el documento como un PDF A4 en w (CA-39.1). ahora es el momento en que se genera: su fecha en la
// Argentina va al pie de cada hoja.
func Escribir(w io.Writer, d Documento, ahora time.Time) error {
	pdf, _, err := generar(d, ahora, true)
	if err != nil {
		return err
	}
	_, err = w.Write(pdf)
	return err
}

// generar arma el PDF en memoria y devuelve también la cantidad de hojas. Los tests lo piden sin comprimir para poder
// buscar los textos.
func generar(d Documento, ahora time.Time, comprimir bool) ([]byte, int, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetCompression(comprimir)
	pdf.SetMargins(margen, margen, margen)
	pdf.SetAutoPageBreak(true, margen+alturaPie)
	if d.Titulo != "" {
		pdf.SetTitle(d.Titulo, true)
	}
	pdf.AliasNbPages("")
	generadoEl := domain.HoyEnArgentina(ahora).Format(formatoFecha)
	pdf.SetFooterFunc(func() {
		pdf.SetY(-alturaPie - 3)
		pdf.SetFont("Helvetica", "", 8)
		pdf.SetTextColor(110, 110, 110)
		pdf.CellFormat(0, alturaLinea, cp1252(fmt.Sprintf("Generado el %s · Página %d de {nb}", generadoEl, pdf.PageNo())), "", 0, "C", false, 0, "")
	})
	pdf.AddPage()

	e := escritor{pdf: pdf}
	for _, b := range d.Bloques {
		e.bloque(b)
	}

	var salida bytes.Buffer
	if err := pdf.Output(&salida); err != nil {
		return nil, 0, fmt.Errorf("generar el PDF: %w", err)
	}
	return salida.Bytes(), pdf.PageCount(), nil
}

// escritor dibuja los bloques uno debajo del otro.
type escritor struct {
	pdf *fpdf.Fpdf
}

func (e escritor) ancho() float64 {
	ancho, _ := e.pdf.GetPageSize()
	return ancho - 2*margen
}

// limite es hasta dónde se puede escribir antes del pie.
func (e escritor) limite() float64 {
	_, alto := e.pdf.GetPageSize()
	return alto - margen - alturaPie
}

// lugarPara pasa a una hoja nueva si lo que sigue no entra en lo que queda de esta.
func (e escritor) lugarPara(alto float64) {
	if e.pdf.GetY()+alto > e.limite() {
		e.pdf.AddPage()
	}
}

func (e escritor) texto(estilo string, tamano float64, gris int) {
	e.pdf.SetFont("Helvetica", estilo, tamano)
	e.pdf.SetTextColor(gris, gris, gris)
}

func (e escritor) bloque(b Bloque) {
	pdf := e.pdf
	switch b.Tipo {
	case Titulo:
		e.texto("B", 16, 20)
		pdf.MultiCell(0, 8, cp1252(b.Texto), "", "L", false)
		pdf.Ln(1)
	case Subtitulo:
		e.lugarPara(30)
		e.texto("B", 13, 20)
		pdf.MultiCell(0, 7, cp1252(b.Texto), "", "L", false)
	case Seccion:
		e.lugarPara(25) // que el título no quede solo al final de la hoja
		pdf.Ln(3)
		e.texto("B", 11, 20)
		pdf.MultiCell(0, 6, cp1252(b.Texto), "", "L", false)
		pdf.Ln(1)
	case Parrafo:
		e.texto("", 10, 30)
		pdf.MultiCell(0, alturaLinea, cp1252(b.Texto), "", "L", false)
	case Detalle:
		e.texto("", 9, 110)
		pdf.MultiCell(0, alturaLinea, cp1252(b.Texto), "", "L", false)
		pdf.Ln(2)
	case Aviso:
		e.texto("", 10, 30)
		pdf.SetFillColor(220, 236, 248)
		pdf.MultiCell(0, 7, cp1252(b.Texto), "", "L", true)
		pdf.Ln(2)
	case Tabla:
		e.tabla(b)
	case SaltoDeHoja:
		pdf.AddPage()
	}
}

// tabla dibuja los encabezados y las filas. Cada celda parte su texto en líneas (RN5) y la fila toma el alto de la
// celda más alta. Si una fila no entra, sigue en la hoja siguiente con los encabezados otra vez.
func (e escritor) tabla(b Bloque) {
	anchos := e.anchos(b.Anchos)
	e.fila(b.Encabezados, anchos, true)
	for _, f := range b.Filas {
		e.texto("", 9, 30)
		if e.pdf.GetY()+e.altoFila(f, anchos) > e.limite() {
			e.pdf.AddPage()
			e.fila(b.Encabezados, anchos, true)
			e.texto("", 9, 30)
		}
		e.fila(f, anchos, false)
	}
	e.pdf.Ln(2)
}

// anchos reparte el ancho de la hoja según las proporciones de cada columna.
func (e escritor) anchos(proporciones []float64) []float64 {
	total := 0.0
	for _, p := range proporciones {
		total += p
	}
	anchos := make([]float64, len(proporciones))
	for i, p := range proporciones {
		anchos[i] = e.ancho() * p / total
	}
	return anchos
}

func (e escritor) lineas(texto string, ancho float64) [][]byte {
	lineas := e.pdf.SplitLines([]byte(cp1252(texto)), ancho)
	if len(lineas) == 0 {
		return [][]byte{nil}
	}
	return lineas
}

func (e escritor) altoFila(celdas []string, anchos []float64) float64 {
	maximo := 1
	for i, c := range celdas {
		maximo = max(maximo, len(e.lineas(c, anchos[i])))
	}
	return float64(maximo)*alturaLinea + 2*relleno
}

func (e escritor) fila(celdas []string, anchos []float64, encabezado bool) {
	pdf := e.pdf
	estilo := ""
	if encabezado {
		e.texto("B", 9, 20)
		pdf.SetFillColor(233, 236, 239)
		estilo = "F"
	}
	pdf.SetDrawColor(200, 200, 200)
	alto := e.altoFila(celdas, anchos)
	x, y := margen, pdf.GetY()
	for i, c := range celdas {
		pdf.Rect(x, y, anchos[i], alto, estilo+"D")
		for n, linea := range e.lineas(c, anchos[i]) {
			pdf.SetXY(x, y+relleno+float64(n)*alturaLinea)
			pdf.CellFormat(anchos[i], alturaLinea, string(linea), "", 0, "L", false, 0, "")
		}
		x += anchos[i]
	}
	pdf.SetXY(margen, y+alto)
}

// cp1252 pasa el texto a Windows-1252, la codificación de las fuentes que trae fpdf. Lo que no existe en ella (emojis,
// flechas, otros alfabetos) queda como "?" (RN4).
func cp1252(texto string) string {
	var b strings.Builder
	for _, r := range texto {
		if c, ok := charmap.Windows1252.EncodeRune(r); ok {
			b.WriteByte(c)
		} else {
			b.WriteByte('?')
		}
	}
	return b.String()
}
