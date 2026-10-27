package server_test

import (
	"context"
	"net/http"
	"net/url"
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/poker"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

// Mensajes de la spec de HU-19 (secciones 3 y 7).
const (
	mensajeRondaRevelada = "Ronda revelada."
	botonCerrarRonda     = "Cerrar la ronda y revelar"
	rutaCerrarRonda      = rutaPoker + "/cerrar"
)

// rondasConVotos es el repositorio de HU-19: Revelar del store en memoria y los votos de cada ronda, que guarda HU-18.
type rondasConVotos struct {
	*memory.Poker
	votos map[int64][]poker.Voto
}

func (r *rondasConVotos) Votos(_ context.Context, rondaID int64) ([]poker.Voto, error) {
	return r.votos[rondaID], nil
}

// nuevaAppRevelar es la aplicación de HU-17 con la sesión de HU-1 abierta por Juan y estos votos en la ronda 1.
func nuevaAppRevelar(t *testing.T, votos map[string]int) (*appPoker, http.Handler) {
	t.Helper()
	sesiones := memory.NuevoPoker()
	a := nuevaAppPoker(t, "", sesiones)
	v, err := a.dep.Poker.Abrir(context.Background(), 1, 1, a.porNombre["Juan"].ID)
	if err != nil {
		t.Fatalf("abrir la sesión: %v", err)
	}
	repo := &rondasConVotos{Poker: sesiones, votos: map[int64][]poker.Voto{}}
	for nombre, valor := range votos {
		repo.votos[v.Ronda.ID] = append(repo.votos[v.Ronda.ID], poker.Voto{RondaID: v.Ronda.ID, IntegranteID: a.porNombre[nombre].ID, Valor: valor})
	}
	a.dep.RevelarPoker = service.NuevoRevelarPoker(a.dep.Poker, repo)
	return a, a.sinLogin()
}

// RN6 · falta votar alguien: no se ven los valores y se ofrece cerrar la ronda
func TestRevelar_SinRevelar_OfreceCerrarLaRonda(t *testing.T) {
	_, h := nuevaAppRevelar(t, map[string]int{"Carolina": 8})

	rec := pedido(h, rutaPoker, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("código = %d, se esperaba 200", rec.Code)
	}
	cuerpo := rec.Body.String()
	verificarContiene(t, cuerpo, botonCerrarRonda, `action="`+rutaCerrarRonda+`"`, "Quién cierra")
	verificarNoContiene(t, cuerpo, "Votos de la ronda", mensajeRondaRevelada)
}

// CA-19.1, RN1 y CA-19.2 · votaron todos: al ver la sesión la ronda se revela sola y muestra cada voto con su autor
func TestRevelar_VotaronTodos_SeRevelaSola(t *testing.T) {
	_, h := nuevaAppRevelar(t, map[string]int{"Juan": 5, "Carolina": 8})

	cuerpo := pedido(h, rutaPoker, nil).Body.String()

	verificarContiene(t, cuerpo, mensajeRondaRevelada, "Votos de la ronda 1", "Carolina", "<td>8</td>", "<td>5</td>")
	verificarNoContiene(t, cuerpo, botonCerrarRonda)
}

// CA-19.1, RN2 y CA-19.2 · el Agile Enabler cierra la ronda: 303 a la sesión, que muestra los votos y quién no votó
func TestRevelar_ElAgileEnablerCierraLaRonda(t *testing.T) {
	a, h := nuevaAppRevelar(t, map[string]int{"Carolina": 8})

	rec := pedido(h, rutaCerrarRonda, a.abre("Juan"))

	if rec.Code != http.StatusSeeOther || rec.Header().Get("Location") != rutaPoker+"?revelada=1" {
		t.Fatalf("código = %d, Location = %q; se esperaba 303 a %s?revelada=1", rec.Code, rec.Header().Get("Location"), rutaPoker)
	}
	cuerpo := pedido(h, rutaPoker+"?revelada=1", nil).Body.String()
	verificarContiene(t, cuerpo, mensajeRondaRevelada, "Votos de la ronda 1", "<td>8</td>", "no votó")
	verificarNoContiene(t, cuerpo, botonCerrarRonda)
}

// CA-19.3, RN4 y RN5 · cuándo no se cierra la ronda: el código y el aviso de la spec (sección 7)
func TestRevelar_CerrarConError(t *testing.T) {
	casos := []struct {
		nombre string
		votos  map[string]int
		quien  string
		codigo int
		texto  string
	}{
		{"un Product Builder", map[string]int{"Carolina": 8}, "Carolina", http.StatusForbidden, "solo el Agile Enabler del proyecto"},
		{"sin elegir quién", map[string]int{"Carolina": 8}, "", http.StatusUnprocessableEntity, "solo el Agile Enabler del proyecto"},
		{"la ronda sin votos", nil, "Juan", http.StatusConflict, "la ronda no tiene votos: no se puede revelar"},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			a, h := nuevaAppRevelar(t, c.votos)
			campos := url.Values{"quien": {""}}
			if c.quien != "" {
				campos = a.abre(c.quien)
			}

			rec := pedido(h, rutaCerrarRonda, campos)

			if rec.Code != c.codigo {
				t.Fatalf("código = %d, se esperaba %d", rec.Code, c.codigo)
			}
			verificarContiene(t, rec.Body.String(), c.texto, botonCerrarRonda)
			verificarNoContiene(t, rec.Body.String(), "Votos de la ronda")
		})
	}
}

// RN5 · una ronda ya revelada no se vuelve a cerrar
func TestRevelar_CerrarUnaRondaYaRevelada(t *testing.T) {
	a, h := nuevaAppRevelar(t, map[string]int{"Carolina": 8})
	pedido(h, rutaCerrarRonda, a.abre("Juan"))

	rec := pedido(h, rutaCerrarRonda, a.abre("Juan"))

	if rec.Code != http.StatusConflict {
		t.Fatalf("código = %d, se esperaba 409", rec.Code)
	}
	verificarContiene(t, rec.Body.String(), "la ronda ya está revelada", "Votos de la ronda 1")
}
