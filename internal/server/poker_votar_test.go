package server_test

import (
	"context"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/member"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/poker"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

// Mensajes de la spec de HU-18 (secciones 3 y 7).
const (
	mensajeVotoRegistrado = "Voto registrado. Podés cambiarlo hasta que se revele la ronda."
	mensajeNoSeGuardoVoto = "No se pudo guardar el voto. Probá de nuevo en unos minutos."
)

const rutaVotos = rutaPoker + "/votos"

// conSesionAbierta abre la sesión de HU-1 Login con el Agile Enabler (Juan, o Ana si es la Agile Enabler).
func (a *appPoker) conSesionAbierta(t *testing.T) *appPoker {
	t.Helper()
	ctx := context.Background()
	quien := a.porNombre["Juan"]
	if ana, ok := a.porNombre["Ana"]; ok && ana.Rol == member.RolAgileEnabler {
		quien = ana
	}
	if _, err := a.dep.Poker.Abrir(ctx, 1, 1, quien.ID); err != nil {
		t.Fatalf("abrir la sesión: %v", err)
	}
	return a
}

func (a *appPoker) vota(nombre, valor string) url.Values {
	campos := a.abre(nombre)
	campos.Set("valor", valor)
	return campos
}

// estadoDeVoto es como se ve un participante en la lista de quién votó.
func estadoDeVoto(nombre string, voto bool) string {
	if voto {
		return `<span>` + nombre + `</span><span class="badge text-bg-success">Votó</span>`
	}
	return `<span>` + nombre + `</span><span class="badge text-bg-secondary">Falta votar</span>`
}

// T5 · con la sesión abierta, la página muestra las cartas de la escala, Quién vota (sin login) y quién votó, que se
// actualiza sola cada 5 segundos con HTMX
func TestPoker_SesionAbierta_MuestraLasCartasYQuienVoto(t *testing.T) {
	rec := pedido(nuevaAppPoker(t, "", nil).conSesionAbierta(t).sinLogin(), rutaPoker, nil)

	cuerpo := rec.Body.String()
	verificarContiene(t, cuerpo, `action="`+rutaVotos+`"`, `name="quien"`, "Quién vota",
		`hx-get="`+rutaVotos+`"`, `hx-trigger="every 5s"`, "0 de 2 votaron", estadoDeVoto("Juan", false), estadoDeVoto("Carolina", false))
	for _, valor := range []string{"0", "1", "2", "3", "5", "8", "13", "21"} {
		verificarContiene(t, cuerpo, `name="valor" value="`+valor+`"`)
	}
}

// Sin sesión abierta no hay cartas
func TestPoker_SinSesion_NoMuestraLasCartas(t *testing.T) {
	rec := pedido(nuevaAppPoker(t, "", nil).sinLogin(), rutaPoker, nil)

	verificarNoContiene(t, rec.Body.String(), `name="valor"`, "votaron")
}

// CA-18.1 y CA-18.2 · votar vuelve a la página (303) con el aviso, y se ve quién votó
func TestPoker_VotarSinLogin(t *testing.T) {
	a := nuevaAppPoker(t, "", nil).conSesionAbierta(t)
	h := a.sinLogin()

	rec := pedido(h, rutaVotos, a.vota("Carolina", "5"))

	if rec.Code != http.StatusSeeOther || rec.Header().Get("Location") != rutaPoker+"?votado=1" {
		t.Fatalf("código = %d, Location = %q; se esperaba 303 a %s?votado=1", rec.Code, rec.Header().Get("Location"), rutaPoker)
	}
	cuerpo := pedido(h, rutaPoker+"?votado=1", nil).Body.String()
	verificarContiene(t, cuerpo, mensajeVotoRegistrado, "1 de 2 votaron", estadoDeVoto("Carolina", true), estadoDeVoto("Juan", false))
}

// CA-18.2, CA-18.3 y RN5 · ni la página ni el fragmento de HTMX tienen los valores, ni siquiera el propio: ninguna carta
// queda marcada y el fragmento no tiene números de la escala
func TestPoker_NadieVeLosValoresAntesDeRevelar(t *testing.T) {
	a := nuevaAppPoker(t, "", nil).conSesionAbierta(t)
	h := a.sinLogin()
	pedido(h, rutaVotos, a.vota("Carolina", "13"))
	pedido(h, rutaVotos, a.vota("Juan", "21"))

	pagina := pedido(h, rutaPoker, nil).Body.String()
	fragmento := pedido(h, rutaVotos, nil, porHTMX).Body.String()

	verificarNoContiene(t, pagina, `aria-pressed="true"`, "checked")
	verificarContiene(t, fragmento, "2 de 2 votaron")
	sinRuta := strings.ReplaceAll(fragmento, rutaVotos, "")
	verificarNoContiene(t, sinRuta, "13", "21")
}

// T5 · GET …/poker/votos devuelve solo el fragmento de quién votó, para que HTMX lo reemplace
func TestPoker_FragmentoDeQuienVoto(t *testing.T) {
	a := nuevaAppPoker(t, "", nil).conSesionAbierta(t)
	h := a.sinLogin()
	pedido(h, rutaVotos, a.vota("Juan", "3"))

	rec := pedido(h, rutaVotos, nil, porHTMX)

	if rec.Code != http.StatusOK {
		t.Fatalf("código = %d, se esperaba 200", rec.Code)
	}
	cuerpo := rec.Body.String()
	verificarContiene(t, cuerpo, `id="quien-voto"`, `hx-get="`+rutaVotos+`"`, "1 de 2 votaron", estadoDeVoto("Juan", true))
	verificarNoContiene(t, cuerpo, "<html", "<nav", `name="valor"`)
}

// porHTMX marca el pedido como lo hace HTMX.
func porHTMX(r *http.Request) { r.Header.Set("HX-Request", "true") }

// Después de un error al votar, la dirección queda en …/poker/votos: si se recarga (sin HTMX), vuelve a la página de
// la sesión en lugar de mostrar el fragmento suelto
func TestPoker_FragmentoSinHTMX_VuelveALaPagina(t *testing.T) {
	rec := pedido(nuevaAppPoker(t, "", nil).conSesionAbierta(t).sinLogin(), rutaVotos, nil)

	if rec.Code != http.StatusSeeOther || rec.Header().Get("Location") != rutaPoker {
		t.Errorf("código = %d, Location = %q; se esperaba 303 a %s", rec.Code, rec.Header().Get("Location"), rutaPoker)
	}
}

// Sin sesión abierta el fragmento queda vacío: HTMX lo saca y deja de preguntar
func TestPoker_FragmentoSinSesion(t *testing.T) {
	rec := pedido(nuevaAppPoker(t, "", nil).sinLogin(), rutaVotos, nil, porHTMX)

	if rec.Code != http.StatusOK || strings.TrimSpace(rec.Body.String()) != "" {
		t.Errorf("código = %d, cuerpo = %q; se esperaba 200 sin contenido", rec.Code, rec.Body.String())
	}
}

// CA-18.4, RN2 y RN3 · errores al votar sin login: el código de la spec y el aviso o el error en su campo
func TestPoker_VotarSinLogin_Errores(t *testing.T) {
	casos := []struct {
		nombre string
		ruta   string
		quien  string
		valor  string
		codigo int
		texto  string
	}{
		{"fuera de la escala (CA-18.4)", rutaVotos, "Carolina", "4", http.StatusUnprocessableEntity, `<div class="invalid-feedback d-block">` + poker.ErrVotoFueraDeEscala.Error()},
		{"sin valor (CA-18.4)", rutaVotos, "Carolina", "", http.StatusUnprocessableEntity, poker.ErrVotoFueraDeEscala.Error()},
		{"no se eligió quién vota (RN2)", rutaVotos, "", "5", http.StatusUnprocessableEntity, `<div class="invalid-feedback">` + poker.ErrSoloParticipantes.Error()},
		{"una integrante dada de baja (RN2)", rutaVotos, "Lucía", "5", http.StatusForbidden, poker.ErrSoloParticipantes.Error()},
		{"historia sin sesión (RN3)", "/proyectos/1/backlog/2/poker/votos", "Carolina", "5", http.StatusConflict, poker.ErrSesionNoEncontrada.Error()},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			a := nuevaAppPoker(t, "", nil).conSesionAbierta(t)
			campos := url.Values{"quien": {""}, "valor": {c.valor}}
			if c.quien != "" {
				campos = a.vota(c.quien, c.valor)
			}

			rec := pedido(a.sinLogin(), c.ruta, campos)

			if rec.Code != c.codigo {
				t.Errorf("código = %d, se esperaba %d", rec.Code, c.codigo)
			}
			verificarContiene(t, rec.Body.String(), c.texto)
			verificarContiene(t, pedido(a.sinLogin(), rutaVotos, nil, porHTMX).Body.String(), "0 de 2 votaron")
		})
	}
}

// pokerRevelado es el repositorio en memoria con la ronda actual ya revelada (HU-19 todavía no la revela).
type pokerRevelado struct{ *memory.Poker }

func (p pokerRevelado) SesionAbierta(ctx context.Context, historiaID int64) (poker.Sesion, poker.Ronda, error) {
	s, r, err := p.Poker.SesionAbierta(ctx, historiaID)
	r.Revelada = true
	return s, r, err
}

// RN3 · con la ronda revelada no se vota (409) y no hay cartas
func TestPoker_RondaRevelada(t *testing.T) {
	a := nuevaAppPoker(t, "", pokerRevelado{memory.NuevoPoker()}).conSesionAbierta(t)

	rec := pedido(a.sinLogin(), rutaVotos, a.vota("Carolina", "5"))

	if rec.Code != http.StatusConflict {
		t.Errorf("código = %d, se esperaba 409", rec.Code)
	}
	verificarContiene(t, rec.Body.String(), poker.ErrRondaRevelada.Error())
	verificarNoContiene(t, rec.Body.String(), `name="valor"`)
}

// pokerSinVotos es el repositorio en memoria con la base caída al guardar el voto.
type pokerSinVotos struct{ *memory.Poker }

func (pokerSinVotos) Votar(context.Context, int64, int64, int) error { return errPokerCaido }

func TestPoker_Votar_FallaLaBase(t *testing.T) {
	a := nuevaAppPoker(t, "", pokerSinVotos{memory.NuevoPoker()}).conSesionAbierta(t)

	rec := pedido(a.sinLogin(), rutaVotos, a.vota("Carolina", "5"))

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("código = %d, se esperaba 500", rec.Code)
	}
	verificarContiene(t, rec.Body.String(), mensajeNoSeGuardoVoto)
}

// Si falla la base al ver quién votó, 500 con el mensaje de HU-17
func TestPoker_FragmentoFallaLaBase(t *testing.T) {
	rec := pedido(nuevaAppPoker(t, "", pokerQueFalla{}).sinLogin(), rutaVotos, nil, porHTMX)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("código = %d, se esperaba 500", rec.Code)
	}
}

func TestPoker_Votar_NoEncontrado(t *testing.T) {
	a := nuevaAppPoker(t, "", nil)
	for _, ruta := range []string{"/proyectos/9/backlog/1/poker/votos", "/proyectos/1/backlog/9/poker/votos", "/proyectos/1/backlog/x/poker/votos"} {
		if rec := pedido(a.sinLogin(), ruta, nil, porHTMX); rec.Code != http.StatusNotFound {
			t.Errorf("GET %s: código = %d, se esperaba 404", ruta, rec.Code)
		}
		if rec := pedido(a.sinLogin(), ruta, a.vota("Juan", "5")); rec.Code != http.StatusNotFound {
			t.Errorf("POST %s: código = %d, se esperaba 404", ruta, rec.Code)
		}
	}
}

// RN6 · con login vota el integrante de la cuenta: no se elige quién vota (y si viene, no se usa)
func TestPoker_ConLogin_VotaElIntegranteDeLaCuenta(t *testing.T) {
	a := nuevaAppPoker(t, member.RolProductBuilder, nil).conSesionAbierta(t)
	h := a.conLogin()

	pagina := pedido(h, rutaPoker, nil, conSesion("token-de-ana")).Body.String()
	verificarNoContiene(t, pagina, `name="quien"`)
	verificarContiene(t, pagina, `name="valor" value="5"`)

	rec := pedido(h, rutaVotos, a.vota("Juan", "5"), conSesion("token-de-ana"))
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("código = %d, se esperaba 303: %s", rec.Code, rec.Body.String())
	}
	fragmento := pedido(h, rutaVotos, nil, conSesion("token-de-ana"), porHTMX).Body.String()
	verificarContiene(t, fragmento, estadoDeVoto("Ana", true), estadoDeVoto("Juan", false))
}

// Con login, una cuenta que no es integrante del proyecto no vota
func TestPoker_ConLogin_NoEsIntegranteNoVota(t *testing.T) {
	a := nuevaAppPoker(t, "", nil).conSesionAbierta(t)

	rec := pedido(a.conLogin(), rutaVotos, url.Values{"valor": {"5"}}, conSesion("token-de-ana"))

	if rec.Code != http.StatusForbidden {
		t.Errorf("código = %d, se esperaba 403", rec.Code)
	}
	verificarContiene(t, rec.Body.String(), mensajeNoIntegrante)
}

// TEC-05 · con el login activado, votar y ver quién votó piden sesión
func TestPoker_ConLogin_VotarPideSesion(t *testing.T) {
	h := nuevaAppPoker(t, "", nil).conSesionAbierta(t).conLogin()
	if rec := pedido(h, rutaVotos, url.Values{"valor": {"5"}}); rec.Code != http.StatusUnauthorized {
		t.Errorf("votar sin sesión: código = %d, se esperaba 401", rec.Code)
	}
	if rec := pedido(h, rutaVotos, nil, porHTMX); rec.Code == http.StatusOK {
		t.Errorf("ver quién votó sin sesión: código = %d, no se esperaba 200", rec.Code)
	}
}
