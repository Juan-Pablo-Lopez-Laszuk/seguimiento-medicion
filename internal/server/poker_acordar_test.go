package server_test

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/member"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/poker"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/store/memory"
)

// Mensajes de la spec de HU-22 (secciones 3 y 7).
const (
	mensajeAcordada      = "Estimación acordada: 5 SP. La sesión de Planning Poker quedó cerrada."
	mensajeSinRevelar    = "Para registrar la estimación, primero revelá la ronda."
	mensajeSinCerradas   = "Todavía no se cerró ninguna sesión."
	mensajeNoSeRegistro  = "No se pudo registrar la estimación. Probá de nuevo en unos minutos."
	mensajeValorInvalido = "los story points tienen que ser 0, 1, 2, 3, 5, 8, 13 o 21"
)

const rutaAcordar = rutaPoker + "/acordar"

// pokerRevelable es el repositorio en memoria con la ronda actual revelada o no, a pedido: revelar es de HU-19. Si
// fallaCerrar, cerrar responde como una base caída.
type pokerRevelable struct {
	*memory.Poker
	revelada    bool
	fallaCerrar bool
}

func (p *pokerRevelable) SesionAbierta(ctx context.Context, historiaID int64) (poker.Sesion, poker.Ronda, error) {
	s, r, err := p.Poker.SesionAbierta(ctx, historiaID)
	r.Revelada = r.Revelada || p.revelada
	return s, r, err
}

func (p *pokerRevelable) Cerrar(ctx context.Context, sesionID, historiaID int64, valor int) error {
	if p.fallaCerrar {
		return errPokerCaido
	}
	return p.Poker.Cerrar(ctx, sesionID, historiaID, valor)
}

// nuevaAppAcordar es la aplicación de HU-17 con el caso de uso de HU-22 conectado. Si conSesion, quien conduce (Juan,
// o Ana si es la Agile Enabler) ya abrió la sesión de HU-1 Login.
func nuevaAppAcordar(t *testing.T, anaRol member.Rol, revelada, conSesionAbierta bool) (*appPoker, *pokerRevelable) {
	t.Helper()
	repo := &pokerRevelable{Poker: memory.NuevoPoker(), revelada: revelada}
	a := nuevaAppPoker(t, anaRol, repo)
	repo.ConHistorias(a.historias)
	a.dep.Poker.ConAcordar(repo)
	if conSesionAbierta {
		conduce := a.porNombre["Juan"].ID
		if anaRol == member.RolAgileEnabler {
			conduce = a.porNombre["Ana"].ID
		}
		if _, err := a.dep.Poker.Abrir(context.Background(), 1, 1, conduce); err != nil {
			t.Fatalf("abrir la sesión: %v", err)
		}
	}
	return a, repo
}

func (a *appPoker) acuerda(nombre, valor string) url.Values {
	campos := url.Values{"valor": {valor}}
	if nombre != "" {
		campos.Set("quien", a.abre(nombre).Get("quien"))
	}
	return campos
}

// RN4 · con la ronda revelada la sesión ofrece registrar el valor: la escala y, sin login, quién registra
func TestPoker_ConLaRondaRevelada_OfreceRegistrar(t *testing.T) {
	a, _ := nuevaAppAcordar(t, "", true, true)

	cuerpo := pedido(a.sinLogin(), rutaPoker, nil).Body.String()

	verificarContiene(t, cuerpo, `action="`+rutaAcordar+`"`, "Valor acordado", `name="valor"`, "Quién registra", "Registrar estimación")
	for _, valor := range backlog.EscalaStoryPoints {
		verificarContiene(t, cuerpo, fmt.Sprintf(`<option value="%d">`, valor))
	}
	verificarNoContiene(t, cuerpo, mensajeSinRevelar)
}

// RN4 · sin revelar la ronda no se ofrece el formulario
func TestPoker_SinRevelar_NoOfreceRegistrar(t *testing.T) {
	a, _ := nuevaAppAcordar(t, "", false, true)

	cuerpo := pedido(a.sinLogin(), rutaPoker, nil).Body.String()

	verificarContiene(t, cuerpo, mensajeSinRevelar)
	verificarNoContiene(t, cuerpo, `action="`+rutaAcordar+`"`, "Registrar estimación")
}

// CA-22.2 CA-22.3 · registrar: 303 a la página, que muestra el mensaje, los story points y la sesión cerrada
func TestPoker_AcordarSinLogin(t *testing.T) {
	a, _ := nuevaAppAcordar(t, "", true, true)
	h := a.sinLogin()

	rec := pedido(h, rutaAcordar, a.acuerda("Juan", "5"))

	if rec.Code != http.StatusSeeOther || rec.Header().Get("Location") != rutaPoker+"?acordada=5" {
		t.Fatalf("código = %d, Location = %q; se esperaba 303 a %s?acordada=5: %s", rec.Code, rec.Header().Get("Location"), rutaPoker, rec.Body.String())
	}
	cuerpo := pedido(h, rutaPoker+"?acordada=5", nil).Body.String()
	verificarContiene(t, cuerpo, mensajeAcordada, "Story points actuales: 5", "Sesiones cerradas", "Sesión 1", "Abrir sesión")
	verificarNoContiene(t, cuerpo, "Sesión abierta", mensajeSinCerradas)
}

// CA-22.3 · sin sesiones cerradas la página lo dice
func TestPoker_SinSesionesCerradas(t *testing.T) {
	a, _ := nuevaAppAcordar(t, "", true, false)

	verificarContiene(t, pedido(a.sinLogin(), rutaPoker, nil).Body.String(), "Sesiones cerradas", mensajeSinCerradas)
}

// El mensaje de éxito solo sale de un valor de la escala: lo que venga en la URL no se muestra
func TestPoker_MensajeAcordada_SoloConUnValorDeLaEscala(t *testing.T) {
	a, _ := nuevaAppAcordar(t, "", true, false)
	for _, valor := range []string{"4", "%3Cscript%3E", "", "5x"} {
		cuerpo := pedido(a.sinLogin(), rutaPoker+"?acordada="+valor, nil).Body.String()
		verificarNoContiene(t, cuerpo, "Estimación acordada", "<script>")
	}
}

// Errores al registrar, sin login: el código de la spec y la página con el aviso
func TestPoker_AcordarSinLogin_Errores(t *testing.T) {
	casos := []struct {
		nombre       string
		revelada     bool
		abierta      bool
		hecha        bool
		quien        string
		valor        string
		codigo       int
		texto        string
		sigueAbierta bool
	}{
		{"valor fuera de la escala (CA-22.1)", true, true, false, "Juan", "4", http.StatusUnprocessableEntity, mensajeValorInvalido, true},
		{"lo registra una Product Builder (RN3)", true, true, false, "Carolina", "5", http.StatusForbidden, poker.ErrSoloAgileEnablerAcuerda.Error(), true},
		{"no se eligió quién registra (RN3)", true, true, false, "", "5", http.StatusUnprocessableEntity, poker.ErrSoloAgileEnablerAcuerda.Error(), true},
		{"la ronda sin revelar (RN4)", false, true, false, "Juan", "5", http.StatusConflict, mensajeSinRevelar, true},
		{"sin sesión abierta", true, false, false, "Juan", "5", http.StatusNotFound, poker.ErrSesionNoEncontrada.Error(), false},
		{"la historia pasó a Hecha (RN7)", true, true, true, "Juan", "5", http.StatusConflict, backlog.ErrEstimacionDeHistoriaHecha.Error(), true},
	}
	for _, c := range casos {
		t.Run(c.nombre, func(t *testing.T) {
			a, _ := nuevaAppAcordar(t, "", c.revelada, c.abierta)
			if c.hecha {
				h, _ := a.historias.BuscarPorNumero(context.Background(), 1, 1)
				if err := a.historias.ActualizarEstado(context.Background(), h.ID, backlog.EstadoHecho); err != nil {
					t.Fatal(err)
				}
			}

			rec := pedido(a.sinLogin(), rutaAcordar, a.acuerda(c.quien, c.valor))

			if rec.Code != c.codigo {
				t.Errorf("código = %d, se esperaba %d", rec.Code, c.codigo)
			}
			verificarContiene(t, strings.ToLower(rec.Body.String()), strings.ToLower(c.texto))
			if c.sigueAbierta {
				verificarContiene(t, rec.Body.String(), "Sesión abierta")
			}
		})
	}
}

// Los errores del valor y de quién registra van debajo de su desplegable
func TestPoker_AcordarSinLogin_LosErroresVanEnElCampo(t *testing.T) {
	a, _ := nuevaAppAcordar(t, "", true, true)
	h := a.sinLogin()

	valor := pedido(h, rutaAcordar, a.acuerda("Juan", "4")).Body.String()
	verificarContiene(t, valor, "is-invalid", `<div class="invalid-feedback">`+mensajeValorInvalido)

	quien := pedido(h, rutaAcordar, a.acuerda("", "5")).Body.String()
	verificarContiene(t, quien, `<div class="invalid-feedback">`+poker.ErrSoloAgileEnablerAcuerda.Error())
}

func TestPoker_Acordar_NoEncontrado(t *testing.T) {
	a, _ := nuevaAppAcordar(t, "", true, true)
	for _, ruta := range []string{"/proyectos/9/backlog/1/poker/acordar", "/proyectos/1/backlog/9/poker/acordar", "/proyectos/1/backlog/x/poker/acordar"} {
		if rec := pedido(a.sinLogin(), ruta, a.acuerda("Juan", "5")); rec.Code != http.StatusNotFound {
			t.Errorf("POST %s: código = %d, se esperaba 404", ruta, rec.Code)
		}
	}
}

// RN6 · si falla la base al cerrar, 500 y la sesión sigue abierta
func TestPoker_Acordar_FallaLaBase(t *testing.T) {
	a, repo := nuevaAppAcordar(t, "", true, true)
	repo.fallaCerrar = true

	rec := pedido(a.sinLogin(), rutaAcordar, a.acuerda("Juan", "5"))

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("código = %d, se esperaba 500", rec.Code)
	}
	verificarContiene(t, rec.Body.String(), mensajeNoSeRegistro)
	verificarContiene(t, pedido(a.sinLogin(), rutaPoker, nil).Body.String(), "Sesión abierta")
}

// RN3 · con login registra el integrante de la cuenta: no se elige quién registra
func TestPoker_ConLogin_RegistraElIntegranteDeLaCuenta(t *testing.T) {
	a, _ := nuevaAppAcordar(t, member.RolAgileEnabler, true, true)
	h := a.conLogin()

	pagina := pedido(h, rutaPoker, nil, conSesion("token-de-ana")).Body.String()
	verificarContiene(t, pagina, "Registrar estimación")
	verificarNoContiene(t, pagina, `name="quien"`)

	rec := pedido(h, rutaAcordar, a.acuerda("Carolina", "8"), conSesion("token-de-ana"))
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("código = %d, se esperaba 303: %s", rec.Code, rec.Body.String())
	}
	verificarContiene(t, pedido(h, rutaPoker, nil, conSesion("token-de-ana")).Body.String(), "Story points actuales: 8")
}

// Con login: una Product Builder no registra y una cuenta que no es integrante tampoco; sin sesión, 401
func TestPoker_ConLogin_Acordar_Errores(t *testing.T) {
	// Con Ana Product Builder el proyecto no tiene Agile Enabler activo: la sesión se abre directo en el repositorio.
	pb, repo := nuevaAppAcordar(t, member.RolProductBuilder, true, false)
	login, _ := pb.historias.BuscarPorNumero(context.Background(), 1, 1)
	if _, _, err := repo.Abrir(context.Background(), login.ID); err != nil {
		t.Fatal(err)
	}
	rec := pedido(pb.conLogin(), rutaAcordar, pb.acuerda("Juan", "5"), conSesion("token-de-ana"))
	if rec.Code != http.StatusForbidden {
		t.Errorf("Product Builder: código = %d, se esperaba 403", rec.Code)
	}
	verificarContiene(t, rec.Body.String(), poker.ErrSoloAgileEnablerAcuerda.Error())

	ajena, _ := nuevaAppAcordar(t, "", true, true)
	rec = pedido(ajena.conLogin(), rutaAcordar, ajena.acuerda("Juan", "5"), conSesion("token-de-ana"))
	if rec.Code != http.StatusForbidden {
		t.Errorf("no es integrante: código = %d, se esperaba 403", rec.Code)
	}
	verificarContiene(t, rec.Body.String(), mensajeNoIntegrante)

	if rec := pedido(ajena.conLogin(), rutaAcordar, ajena.acuerda("Juan", "5")); rec.Code != http.StatusUnauthorized {
		t.Errorf("sin sesión: código = %d, se esperaba 401", rec.Code)
	}
}
