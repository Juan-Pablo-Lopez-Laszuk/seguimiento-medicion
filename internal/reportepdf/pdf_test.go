package reportepdf

import (
	"bytes"
	"strings"
	"testing"
	"time"
)

var generado = time.Date(2026, 10, 29, 18, 0, 0, 0, time.UTC) // 15:00 del 29/10 en la Argentina

// sinComprimir genera el PDF sin comprimir, así los tests pueden buscar los textos.
func sinComprimir(t *testing.T, d Documento) (string, int) {
	t.Helper()
	pdf, paginas, err := generar(d, generado, false)
	if err != nil {
		t.Fatal(err)
	}
	return string(pdf), paginas
}

// enElPDF es cómo queda un texto dentro del PDF: en Windows-1252 y con los paréntesis y las barras escapados.
func enElPDF(texto string) string {
	return "(" + strings.NewReplacer(`\`, `\\`, "(", `\(`, ")", `\)`).Replace(cp1252(texto)) + ")"
}

func verificarEnElPDF(t *testing.T, pdf string, textos ...string) {
	t.Helper()
	for _, texto := range textos {
		if !strings.Contains(pdf, enElPDF(texto)) {
			t.Errorf("el PDF no tiene el texto %q", texto)
		}
	}
}

func documentoDePrueba() Documento {
	return Documento{Titulo: "Reporte del Sprint 2 · Software Metrics", Bloques: []Bloque{
		{Tipo: Titulo, Texto: "Reporte del Sprint 2 · Software Metrics"},
		{Tipo: Parrafo, Texto: "Planificar sprints"},
		{Tipo: Detalle, Texto: "Del 19/10/2026 al 25/10/2026 · Cerrado"},
		{Tipo: Aviso, Texto: "El sprint no está cerrado: el reporte muestra cómo va hasta hoy."},
		{Tipo: Seccion, Texto: "Métricas"},
		{Tipo: Tabla, Encabezados: []string{"Métrica", "Valor"}, Filas: [][]string{{"Desviación", "-1 h (-8,3 %)"}}, Anchos: []float64{1, 2}},
		{Tipo: SaltoDeHoja},
		{Tipo: Subtitulo, Texto: "Sprint 3 · Ver métricas"},
	}}
}

// CA-39.1 · Escribir genera un PDF
func TestEscribir_GeneraUnPDF(t *testing.T) {
	var pdf bytes.Buffer
	if err := Escribir(&pdf, documentoDePrueba(), generado); err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(pdf.Bytes(), []byte("%PDF-")) || !bytes.Contains(pdf.Bytes(), []byte("%%EOF")) {
		t.Error("lo que se escribió no es un PDF")
	}
}

// CA-39.2 · el PDF tiene todos los textos del documento, con las tildes y la eñe
func TestEscribir_TieneLosTextos(t *testing.T) {
	pdf, _ := sinComprimir(t, documentoDePrueba())

	verificarEnElPDF(t, pdf, "Reporte del Sprint 2 · Software Metrics", "Planificar sprints",
		"Del 19/10/2026 al 25/10/2026 · Cerrado", "El sprint no está cerrado: el reporte muestra cómo va hasta hoy.",
		"Métricas", "Métrica", "Valor", "Desviación", "-1 h (-8,3 %)", "Sprint 3 · Ver métricas")
	if !strings.Contains(pdf, "/Title ") {
		t.Error("el título del documento tiene que ir en las propiedades del PDF")
	}
}

// Sección 3 · cada hoja lleva al pie la fecha en que se generó y su número
func TestEscribir_PieDeHoja(t *testing.T) {
	pdf, paginas := sinComprimir(t, documentoDePrueba())

	if paginas != 2 {
		t.Fatalf("páginas = %d, se esperaban 2", paginas)
	}
	verificarEnElPDF(t, pdf, "Generado el 29/10/2026 · Página 1 de 2", "Generado el 29/10/2026 · Página 2 de 2")
}

// CA-39.2 (HU-38 sección 3) · un salto de hoja arranca una hoja nueva
func TestEscribir_SaltoDeHoja(t *testing.T) {
	d := documentoDePrueba()
	_, conSalto := sinComprimir(t, d)
	d.Bloques = d.Bloques[:6]
	_, sinSalto := sinComprimir(t, d)

	if conSalto != sinSalto+1 {
		t.Errorf("con el salto tiene %d hojas y sin el salto %d: el salto tiene que sumar una", conSalto, sinSalto)
	}
}

// Sección 3 · una tabla que no entra en una hoja sigue en la siguiente, con sus encabezados
func TestEscribir_TablaLargaSigueEnOtraHoja(t *testing.T) {
	filas := make([][]string, 120)
	for i := range filas {
		filas[i] = []string{"HU-1", "Login"}
	}
	d := Documento{Bloques: []Bloque{{Tipo: Tabla, Encabezados: []string{"ID", "Título"}, Filas: filas, Anchos: []float64{1, 3}}}}

	pdf, paginas := sinComprimir(t, d)

	if paginas < 2 {
		t.Fatalf("120 filas tienen que ocupar más de una hoja; ocuparon %d", paginas)
	}
	if n := strings.Count(pdf, enElPDF("Título")); n != paginas {
		t.Errorf("los encabezados aparecen %d veces en %d hojas: tienen que repetirse en cada una", n, paginas)
	}
}

// RN5 · un texto largo se parte en varias líneas dentro de su celda, sin perder palabras
func TestEscribir_TextoLargoSeParteEnLineas(t *testing.T) {
	largo := strings.Repeat("Como usuario quiero exportar el reporte ", 5) + "final"
	d := Documento{Bloques: []Bloque{{Tipo: Tabla, Encabezados: []string{"ID", "Título"}, Filas: [][]string{{"HU-1", largo}}, Anchos: []float64{1, 3}}}}

	pdf, _ := sinComprimir(t, d)

	if strings.Contains(pdf, enElPDF(largo)) {
		t.Fatal("un título de 200 caracteres no entra en una línea de la celda: se tiene que partir")
	}
	if n := strings.Count(pdf, "exportar"); n != 5 {
		t.Errorf("la palabra \"exportar\" aparece %d veces, se esperaban 5", n)
	}
	if !strings.Contains(pdf, "final") {
		t.Error("se perdió el final del texto")
	}
}

// RN4 · lo que las fuentes del PDF no tienen sale como "?"; el castellano sale completo
func TestEscribir_CaracteresQueNoEstan(t *testing.T) {
	d := Documento{Bloques: []Bloque{
		{Tipo: Parrafo, Texto: "Login 🔐 → OAuth"},
		{Tipo: Parrafo, Texto: "¿Año? ¡Ñandú! «pingüino» · 10 €"},
	}}

	pdf, _ := sinComprimir(t, d)

	if !strings.Contains(pdf, "(Login ? ? OAuth)") {
		t.Error(`"Login 🔐 → OAuth" tiene que salir como "Login ? ? OAuth"`)
	}
	verificarEnElPDF(t, pdf, "¿Año? ¡Ñandú! «pingüino» · 10 €")
	if cp1252("Año") != "A\xf1o" {
		t.Errorf("cp1252(\"Año\") = %q, se esperaba la ñ en Windows-1252", cp1252("Año"))
	}
}
