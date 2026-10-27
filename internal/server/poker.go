package server

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/auth"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/member"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/poker"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/project"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
)

// campoQuien es el campo del formulario donde se elige quién hace la acción cuando no hay login (HU-17, RN6).
const campoQuien = "quien"

// mensajeNoEsIntegrante es el aviso de una cuenta que no es integrante del proyecto (spec de HU-17, sección 7).
const mensajeNoEsIntegrante = "No sos integrante de este proyecto."

// pokerHandlers agrupa los handlers de la sesión de Planning Poker de una historia (épica E4). HU-18, HU-19 y HU-22 le
// agregan los suyos en sus archivos.
type pokerHandlers struct {
	casos *service.Poker
	// cuentas dice quién es el usuario con sesión; es nil si el login no está activado (TEC-05).
	cuentas *service.Cuentas
}

// paginaPoker son los datos de poker.html.
type paginaPoker struct {
	service.VistaSesion
	// ConLogin es true si quién es quién sale de la cuenta (no se ofrece elegirlo).
	ConLogin bool
	QuienID  int64
	Mensaje  string
	Aviso    string
	Errores  map[string]string
	// Escala son los valores de las cartas para votar (HU-18, CA-18.4).
	Escala []int
}

// quienParticipa devuelve el ID del integrante que hace el pedido (RN6): con login, el vinculado a la cuenta con la que
// se entró (Cuentas.Quien); sin login, el elegido en el campo "quien" del formulario (0 si no se eligió o no es un
// número). Con login, si la cuenta no es integrante del proyecto el error es member.ErrNoEsIntegrante. La usan abrir
// (HU-17) y votar (HU-18).
func (h pokerHandlers) quienParticipa(r *http.Request, proyectoID int64) (int64, error) {
	if identidad, ok := auth.IdentidadDe(r.Context()); ok && h.cuentas != nil {
		i, err := h.cuentas.Quien(r.Context(), identidad, proyectoID)
		if err != nil {
			return 0, err
		}
		return i.ID, nil
	}
	id, _ := strconv.ParseInt(r.PostFormValue(campoQuien), 10, 64)
	return id, nil
}

// ver muestra la sesión abierta de la historia o, si no tiene, el formulario para abrirla.
func (h pokerHandlers) ver(w http.ResponseWriter, r *http.Request) {
	proyectoID, numero, ok := idsDeHistoria(w, r)
	if !ok {
		return
	}
	v, err := h.casos.Ver(r.Context(), proyectoID, numero)
	if respondeErrorDePoker(w, err, "ver la sesión de Planning Poker", mensajeNoSeCargo) {
		return
	}
	pagina := h.pagina(v)
	if r.URL.Query().Get("abierta") == "1" {
		pagina.Mensaje = "Sesión de Planning Poker abierta."
	}
	if r.URL.Query().Get("votado") == "1" {
		pagina.Mensaje = "Voto registrado. Podés cambiarlo hasta que se revele la ronda."
	}
	renderPoker(w, http.StatusOK, pagina)
}

// abrir abre la sesión de la historia (HU-17) y vuelve a su página (303). Si no se puede, muestra la página con el
// aviso y el código de la spec (sección 7).
func (h pokerHandlers) abrir(w http.ResponseWriter, r *http.Request) {
	proyectoID, numero, ok := idsDeHistoria(w, r)
	if !ok {
		return
	}
	quienID, err := h.quienParticipa(r, proyectoID)
	if err == nil {
		_, err = h.casos.Abrir(r.Context(), proyectoID, numero, quienID)
	}
	if err == nil {
		http.Redirect(w, r, fmt.Sprintf("/proyectos/%d/backlog/%d/poker?abierta=1", proyectoID, numero), http.StatusSeeOther)
		return
	}

	var codigo int
	pedido := paginaPoker{QuienID: quienID, Errores: map[string]string{}}
	switch {
	case errors.Is(err, poker.ErrHistoriaHecha), errors.Is(err, poker.ErrSesionYaAbierta):
		codigo, pedido.Aviso = http.StatusConflict, err.Error()
	case errors.Is(err, member.ErrNoEsIntegrante):
		codigo, pedido.Aviso = http.StatusForbidden, mensajeNoEsIntegrante
	case errors.Is(err, poker.ErrSoloAgileEnabler) && h.cuentas == nil && quienID == 0:
		codigo, pedido.Errores[campoQuien] = http.StatusUnprocessableEntity, err.Error()
	case errors.Is(err, poker.ErrSoloAgileEnabler):
		codigo, pedido.Aviso = http.StatusForbidden, err.Error()
	default:
		respondeErrorDePoker(w, err, "abrir la sesión de Planning Poker", mensajeNoSeAbrio)
		return
	}
	h.mostrarConAviso(w, r, proyectoID, numero, codigo, pedido)
}

// mostrarConAviso vuelve a mostrar la página de la sesión, con el código, quién hizo el pedido, el aviso y los errores
// de campo de un pedido que no se pudo hacer. La usan abrir (HU-17) y votar (HU-18).
func (h pokerHandlers) mostrarConAviso(w http.ResponseWriter, r *http.Request, proyectoID int64, numero, codigo int, pedido paginaPoker) {
	v, err := h.casos.Ver(r.Context(), proyectoID, numero)
	if respondeErrorDePoker(w, err, "volver a mostrar la sesión de Planning Poker", mensajeNoSeCargo) {
		return
	}
	pagina := h.pagina(v)
	pagina.QuienID, pagina.Aviso, pagina.Errores = pedido.QuienID, pedido.Aviso, pedido.Errores
	renderPoker(w, codigo, pagina)
}

func (h pokerHandlers) pagina(v service.VistaSesion) paginaPoker {
	return paginaPoker{VistaSesion: v, ConLogin: h.cuentas != nil, Errores: map[string]string{}, Escala: backlog.EscalaStoryPoints}
}

func renderPoker(w http.ResponseWriter, codigo int, pagina paginaPoker) {
	render(w, codigo, "poker.html", Pagina{Titulo: "Planning Poker de " + pagina.Historia.Identificador(), Datos: pagina})
}

// Mensajes de una falla de la base (spec de HU-17, sección 7).
const (
	mensajeNoSeAbrio = "No se pudo abrir la sesión. Probá de nuevo en unos minutos."
	mensajeNoSeCargo = "No se pudo cargar la sesión de Planning Poker. Probá de nuevo en unos minutos."
)

// respondeErrorDePoker responde 404 si el proyecto o la historia no existen y 500 con mensaje500 si falla otra cosa.
// Devuelve false si no hubo error.
func respondeErrorDePoker(w http.ResponseWriter, err error, accion, mensaje500 string) bool {
	switch {
	case err == nil:
		return false
	case errors.Is(err, project.ErrNoEncontrado):
		proyectoNoEncontrado(w)
	case errors.Is(err, backlog.ErrHistoriaNoEncontrada):
		historiaNoEncontrada(w)
	default:
		log.Printf("%s: %v", accion, err)
		http.Error(w, mensaje500, http.StatusInternalServerError)
	}
	return true
}
