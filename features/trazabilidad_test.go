package features

import (
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// La matriz de docs/trazabilidad-requisitos.md (TEC-06, CA-06.3) dice, para cada requerimiento de la consigna, qué escenario
// BDD y qué tests lo prueban. Estos tests comprueban que la matriz no miente: que cubra todos los requerimientos de la
// consigna, que cada uno marcado como cubierto tenga al menos un escenario o un test, y que todo lo que nombra exista.

const rutaMatriz = "../docs/trazabilidad-requisitos.md"

// requerimientosDeLaConsigna son los requerimientos mínimos de la consigna del TPI, con el bloque al que pertenecen, tal como
// se escriben en la primera columna de la matriz. Si la matriz deja de nombrar alguno, el test falla.
var requerimientosDeLaConsigna = []string{
	"Proyectos · Crear proyectos",
	"Proyectos · Modificar proyectos",
	"Proyectos · Registrar integrantes",
	"Proyectos · Registrar fecha de inicio y finalización",
	"Proyectos · Consultar el estado de un proyecto",
	"Product Backlog · Identificador",
	"Product Backlog · Título",
	"Product Backlog · Descripción",
	"Product Backlog · Prioridad",
	"Product Backlog · Estado",
	"Product Backlog · Story Points",
	"Product Backlog · Criterios de aceptación",
	"Sprints · Crear Sprints",
	"Sprints · Definir un Sprint Goal",
	"Sprints · Asignar historias a un Sprint",
	"Sprints · Registrar historias completadas",
	"Sprints · Cerrar un Sprint",
	"Sprints · Consultar Sprints anteriores",
	"Estimación · Estimar historias mediante Story Points",
	"Estimación · Registrar las estimaciones individuales (Planning Poker)",
	"Estimación · Mantenerlas ocultas hasta finalizar la votación",
	"Estimación · Mostrar las estimaciones realizadas",
	"Estimación · Detectar diferencias entre estimaciones",
	"Estimación · Realizar nuevas rondas",
	"Estimación · Registrar la estimación acordada",
	"Esfuerzo · Integrante",
	"Esfuerzo · Fecha",
	"Esfuerzo · Actividad realizada",
	"Esfuerzo · Horas trabajadas",
	"Esfuerzo · Comparar el esfuerzo estimado con el real",
	"Defectos · Descripción",
	"Defectos · Severidad",
	"Defectos · Estado",
	"Defectos · Historia relacionada",
	"Defectos · Sprint de detección",
	"Defectos · Sprint de resolución",
	"Métricas · Story Points planificados",
	"Métricas · Story Points completados",
	"Métricas · Velocidad del equipo",
	"Métricas · Horas estimadas",
	"Métricas · Horas reales",
	"Métricas · Desviación entre esfuerzo estimado y real",
	"Métricas · Porcentaje de historias completadas",
	"Métricas · Cantidad de defectos detectados",
	"Métricas · Cantidad de defectos resueltos",
	"Dashboard · Estado del proyecto",
	"Dashboard · Representaciones gráficas de métricas",
	"Reportes · Reporte de un Sprint",
	"Reportes · Reporte de un proyecto",
	"Reportes · Historias planificadas y completadas",
	"Reportes · Estimaciones",
	"Reportes · Esfuerzo registrado",
	"Reportes · Métricas",
	"Reportes · Defectos",
	"Reportes · Exportar el informe a PDF",
	"TDD · Cálculo de métricas",
	"TDD · Cálculos de estimación",
	"TDD · Reglas de negocio",
	"TDD · Validaciones",
}

// filaDeMatriz es una fila de las tablas de la matriz: el requerimiento, su historia, los escenarios, los tests y el estado.
type filaDeMatriz struct {
	linea                                              int
	requerimiento, historia, escenarios, tests, estado string
}

// leerMatriz devuelve las filas de datos de todas las tablas de la matriz (las de cinco columnas, sin el encabezado).
func leerMatriz(t *testing.T) []filaDeMatriz {
	t.Helper()
	contenido, err := os.ReadFile(rutaMatriz)
	if err != nil {
		t.Fatalf("no se pudo leer la matriz de trazabilidad: %v", err)
	}
	var filas []filaDeMatriz
	for i, linea := range strings.Split(strings.ReplaceAll(string(contenido), "\r\n", "\n"), "\n") {
		linea = strings.TrimSpace(linea)
		if !strings.HasPrefix(linea, "|") || strings.HasPrefix(linea, "|---") {
			continue
		}
		celdas := strings.Split(strings.Trim(linea, "|"), "|")
		if len(celdas) != 5 || strings.TrimSpace(celdas[0]) == "Requerimiento" {
			continue
		}
		for j := range celdas {
			celdas[j] = strings.TrimSpace(celdas[j])
		}
		filas = append(filas, filaDeMatriz{i + 1, celdas[0], celdas[1], celdas[2], celdas[3], celdas[4]})
	}
	if len(filas) == 0 {
		t.Fatal("la matriz no tiene ninguna fila de requerimientos")
	}
	return filas
}

// testsDelRepositorio devuelve el nombre de todas las funciones de test (Test...) de los archivos *_test.go del repositorio.
func testsDelRepositorio(t *testing.T) []string {
	t.Helper()
	definicion := regexp.MustCompile(`(?m)^func (Test\w+)\(`)
	var nombres []string
	err := filepath.WalkDir("..", func(ruta string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() && ruta != ".." && strings.HasPrefix(d.Name(), ".") {
			return filepath.SkipDir
		}
		if d.IsDir() || !strings.HasSuffix(ruta, "_test.go") {
			return nil
		}
		contenido, err := os.ReadFile(ruta)
		if err != nil {
			return err
		}
		for _, m := range definicion.FindAllStringSubmatch(string(contenido), -1) {
			nombres = append(nombres, m[1])
		}
		return nil
	})
	if err != nil {
		t.Fatalf("no se pudieron recorrer los tests del repositorio: %v", err)
	}
	return nombres
}

// existeTest dice si hay un test con ese nombre; con un asterisco al final alcanza con que haya uno que empiece así.
func existeTest(nombres []string, referencia string) bool {
	prefijo, comodin := strings.CutSuffix(referencia, "*")
	for _, n := range nombres {
		if n == referencia || (comodin && strings.HasPrefix(n, prefijo)) {
			return true
		}
	}
	return false
}

// titulosDeEscenarios devuelve los títulos de los escenarios (y esquemas) de un archivo .feature.
func titulosDeEscenarios(t *testing.T, feature string) []string {
	t.Helper()
	contenido, err := os.ReadFile(filepath.Join(".", feature))
	if err != nil {
		t.Fatalf("no se pudo leer %s: %v", feature, err)
	}
	var titulos []string
	for _, linea := range strings.Split(string(contenido), "\n") {
		linea = strings.TrimSpace(linea)
		for _, prefijo := range []string{"Esquema del escenario:", "Escenario:"} {
			if titulo, ok := strings.CutPrefix(linea, prefijo); ok {
				titulos = append(titulos, strings.TrimSpace(titulo))
			}
		}
	}
	return titulos
}

var (
	referenciaAFeature = regexp.MustCompile("`([a-z0-9-]+\\.feature)`")
	referenciaATest    = regexp.MustCompile("`(Test\\w+\\*?)`")
	referenciaAEscena  = regexp.MustCompile(`«([^»]+)»`)
)

// Cada requerimiento de la consigna tiene su fila en la matriz, y ninguno aparece dos veces.
func TestMatrizDeTrazabilidad_NombraTodosLosRequerimientosDeLaConsigna(t *testing.T) {
	filas := leerMatriz(t)
	vistas := map[string]int{}
	for _, f := range filas {
		vistas[f.requerimiento]++
	}
	for _, requerimiento := range requerimientosDeLaConsigna {
		switch vistas[requerimiento] {
		case 0:
			t.Errorf("la matriz no tiene una fila para el requerimiento %q", requerimiento)
		case 1:
		default:
			t.Errorf("el requerimiento %q aparece %d veces en la matriz", requerimiento, vistas[requerimiento])
		}
	}
	conocidos := map[string]bool{}
	for _, r := range requerimientosDeLaConsigna {
		conocidos[r] = true
	}
	for _, f := range filas {
		if !conocidos[f.requerimiento] {
			t.Errorf("línea %d: %q no es un requerimiento de la consigna (si es nuevo, sumalo a la lista de este test)", f.linea, f.requerimiento)
		}
	}
}

// Cada fila dice si el requerimiento está cubierto o pendiente; uno cubierto nombra al menos un escenario o un test, y uno
// pendiente dice qué historia lo va a cubrir.
func TestMatrizDeTrazabilidad_CadaRequerimientoEstaCubiertoOPendienteConSuHistoria(t *testing.T) {
	for _, f := range leerMatriz(t) {
		switch {
		case strings.HasPrefix(f.estado, "Cubierto"):
			if !referenciaAFeature.MatchString(f.escenarios) && !referenciaATest.MatchString(f.tests) {
				t.Errorf("línea %d: %q está cubierto pero no nombra ningún escenario ni ningún test", f.linea, f.requerimiento)
			}
		case strings.HasPrefix(f.estado, "Pendiente"):
			if !strings.Contains(f.historia, "HU-") {
				t.Errorf("línea %d: %q está pendiente pero no dice qué historia lo va a cubrir", f.linea, f.requerimiento)
			}
		default:
			t.Errorf("línea %d: el estado de %q tiene que empezar con Cubierto o Pendiente y dice %q", f.linea, f.requerimiento, f.estado)
		}
	}
}

// Todo lo que la matriz nombra existe: los archivos .feature, los títulos de escenarios dentro de esos archivos y los tests.
func TestMatrizDeTrazabilidad_TodoLoQueNombraExiste(t *testing.T) {
	nombresDeTests := testsDelRepositorio(t)
	for _, f := range leerMatriz(t) {
		var titulosDeLaFila []string
		for _, m := range referenciaAFeature.FindAllStringSubmatch(f.escenarios, -1) {
			if _, err := os.Stat(m[1]); err != nil {
				t.Errorf("línea %d: el archivo features/%s no existe", f.linea, m[1])
				continue
			}
			titulosDeLaFila = append(titulosDeLaFila, titulosDeEscenarios(t, m[1])...)
		}
		for _, m := range referenciaAEscena.FindAllStringSubmatch(f.escenarios, -1) {
			if !contiene(titulosDeLaFila, m[1]) {
				t.Errorf("línea %d: ningún .feature de la fila tiene el escenario %q", f.linea, m[1])
			}
		}
		for _, m := range referenciaATest.FindAllStringSubmatch(f.tests, -1) {
			if !existeTest(nombresDeTests, m[1]) {
				t.Errorf("línea %d: no existe el test %s", f.linea, m[1])
			}
		}
	}
}

func contiene(lista []string, buscado string) bool {
	for _, e := range lista {
		if e == buscado {
			return true
		}
	}
	return false
}
