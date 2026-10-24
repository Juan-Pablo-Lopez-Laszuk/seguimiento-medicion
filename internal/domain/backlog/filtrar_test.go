package backlog_test

import (
	"slices"
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
)

// backlogParaFiltrar es el de la feature de HU-10, ya en el orden del backlog.
func backlogParaFiltrar() []backlog.Historia {
	con := func(numero int, titulo string, p backlog.Prioridad, e backlog.Estado) backlog.Historia {
		return backlog.Historia{ID: int64(numero), Numero: numero, Titulo: titulo, Prioridad: p, Estado: e}
	}
	return []backlog.Historia{
		con(1, "Iniciar sesión", backlog.PrioridadAlta, backlog.EstadoPendiente),
		con(2, "Registrar usuario", backlog.PrioridadAlta, backlog.EstadoEnSprint),
		con(3, "Editar perfil", backlog.PrioridadMedia, backlog.EstadoEnProgreso),
		con(5, "Reporte de sesiones", backlog.PrioridadMedia, backlog.EstadoPendiente),
		con(4, "Cerrar sesión", backlog.PrioridadBaja, backlog.EstadoHecho),
	}
}

// CA-10.1 y RN1 · cada filtro por separado y combinados; el resultado conserva el orden de la lista (RN3)
func TestFiltrar(t *testing.T) {
	casos := []struct {
		nombre                   string
		estado, prioridad, texto string
		numeros                  []int
	}{
		{"sin filtros", "", "", "", []int{1, 2, 3, 5, 4}},
		{"por estado", "Pendiente", "", "", []int{1, 5}},
		{"por estado Hecho", "Hecho", "", "", []int{4}},
		{"por prioridad", "", "Alta", "", []int{1, 2}},
		{"por texto", "", "", "sesión", []int{1, 5, 4}},
		{"los tres combinados", "Pendiente", "Media", "sesion", []int{5}},
		{"ninguna coincide", "Hecho", "Alta", "", nil},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			got := numeros(backlog.Filtrar(backlogParaFiltrar(), backlog.NuevoFiltro(c.estado, c.prioridad, c.texto)))

			if !slices.Equal(got, c.numeros) {
				t.Errorf("HU %v, se esperaba HU %v", got, c.numeros)
			}
		})
	}
}

// RN2 · la búsqueda no distingue mayúsculas ni tildes, ignora los espacios de los extremos y no confunde ñ con n
func TestFiltrar_BusquedaSinMayusculasNiTildes(t *testing.T) {
	casos := map[string][]int{
		"SESION":        {1, 5, 4},
		"SeSióN":        {1, 5, 4},
		"regístrar":     {2},
		"  perfil  ":    {3},
		"   ":           {1, 2, 3, 5, 4}, // solo espacios: como no buscar
		"usuario nuevo": nil,
	}
	for texto, esperados := range casos {
		got := numeros(backlog.Filtrar(backlogParaFiltrar(), backlog.NuevoFiltro("", "", texto)))
		if !slices.Equal(got, esperados) {
			t.Errorf("texto %q: HU %v, se esperaba HU %v", texto, got, esperados)
		}
	}

	conEnie := []backlog.Historia{{Numero: 1, Titulo: "Diseño"}, {Numero: 2, Titulo: "Diseno"}}
	if got := numeros(backlog.Filtrar(conEnie, backlog.NuevoFiltro("", "", "diseño"))); !slices.Equal(got, []int{1}) {
		t.Errorf("la ñ no se distingue de la n: HU %v, se esperaba HU [1]", got)
	}
}

// RN4 · un estado o una prioridad que no existe se ignora
func TestNuevoFiltro_IgnoraLosValoresQueNoExisten(t *testing.T) {
	f := backlog.NuevoFiltro("Archivada", "Urgente", "")

	if f.Activo() {
		t.Errorf("filtro = %+v, se esperaba que no filtre nada", f)
	}
	if got := len(backlog.Filtrar(backlogParaFiltrar(), f)); got != 5 {
		t.Errorf("se esperaban las 5 historias y se obtuvieron %d", got)
	}
}

// RN5 · el filtro dice si hay algo puesto, para ocultar Subir y Bajar
func TestFiltro_Activo(t *testing.T) {
	casos := map[[3]string]bool{
		{"", "", ""}:          false,
		{"", "", "   "}:       false,
		{"En sprint", "", ""}: true,
		{"", "Baja", ""}:      true,
		{"", "", "perfil"}:    true,
	}
	for valores, activo := range casos {
		f := backlog.NuevoFiltro(valores[0], valores[1], valores[2])
		if f.Activo() != activo {
			t.Errorf("NuevoFiltro%q.Activo() = %v, se esperaba %v", valores, f.Activo(), activo)
		}
	}
}

// El filtro guarda los valores válidos ya limpios, para mostrarlos en el formulario
func TestNuevoFiltro_GuardaLosValoresLimpios(t *testing.T) {
	f := backlog.NuevoFiltro("En progreso", "Media", "  Editar  ")

	if f.Estado != backlog.EstadoEnProgreso || f.Prioridad != backlog.PrioridadMedia || f.Texto != "Editar" {
		t.Errorf("filtro = %+v", f)
	}
}

// Filtrar devuelve una lista nueva: la que recibe queda como estaba
func TestFiltrar_NoCambiaLaListaQueRecibe(t *testing.T) {
	lista := backlogParaFiltrar()

	backlog.Filtrar(lista, backlog.NuevoFiltro("Hecho", "", ""))

	if got := numeros(lista); !slices.Equal(got, []int{1, 2, 3, 5, 4}) {
		t.Errorf("la lista quedó HU %v", got)
	}
}
