package server

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/member"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/poker"
)

// mensajeRondaRevelada es el aviso de una ronda que se acaba de revelar (spec de HU-19, sección 3).
const mensajeRondaRevelada = "Ronda revelada."

// conVotosRevelados completa la página de la sesión con HU-19: si votaron todos, revela la ronda (RN1) y agrega los
// votos de una ronda revelada (CA-19.2). Devuelve true si ya respondió con un error.
func (h pokerHandlers) conVotosRevelados(w http.ResponseWriter, r *http.Request, pagina *paginaPoker) bool {
	if h.revelar == nil || !pagina.Abierta {
		return false
	}
	v, err := h.revelar.Revisar(r.Context(), pagina.Proyecto.ID, pagina.Historia.Numero)
	if respondeErrorDePoker(w, err, "revisar la ronda de Planning Poker", mensajeNoSeCargo) {
		return true
	}
	pagina.Ronda, pagina.VotosRevelados = v.Ronda, v.Votos
	if v.ReveladaAhora || r.URL.Query().Get("revelada") == "1" {
		pagina.Mensaje = mensajeRondaRevelada
	}
	return false
}

// cerrar revela la ronda actual con los votos que haya (HU-19, CA-19.1 y RN2) y vuelve a la sesión (303). Si no se
// puede, muestra la sesión con el aviso y el código de la spec (sección 7).
func (h pokerHandlers) cerrar(w http.ResponseWriter, r *http.Request) {
	proyectoID, numero, ok := idsDeHistoria(w, r)
	if !ok {
		return
	}
	quienID, err := h.quienParticipa(r, proyectoID)
	if err == nil {
		_, err = h.revelar.Cerrar(r.Context(), proyectoID, numero, quienID)
	}
	if err == nil {
		http.Redirect(w, r, fmt.Sprintf("/proyectos/%d/backlog/%d/poker?revelada=1", proyectoID, numero), http.StatusSeeOther)
		return
	}

	var codigo int
	pagina := paginaPoker{QuienID: quienID, Errores: map[string]string{}}
	switch {
	case errors.Is(err, poker.ErrRondaSinVotos), errors.Is(err, poker.ErrRondaYaRevelada), errors.Is(err, poker.ErrSesionNoEncontrada):
		codigo, pagina.Aviso = http.StatusConflict, err.Error()
	case errors.Is(err, member.ErrNoEsIntegrante):
		codigo, pagina.Aviso = http.StatusForbidden, mensajeNoEsIntegrante
	case errors.Is(err, poker.ErrSoloAgileEnabler) && h.cuentas == nil && quienID == 0:
		// El aviso va arriba y no en el campo: la página también tiene el "quien" del formulario de votar (HU-18).
		codigo, pagina.Aviso = http.StatusUnprocessableEntity, err.Error()
	case errors.Is(err, poker.ErrSoloAgileEnabler):
		codigo, pagina.Aviso = http.StatusForbidden, err.Error()
	default:
		respondeErrorDePoker(w, err, "cerrar la ronda de Planning Poker", mensajeNoSeCerro)
		return
	}
	v, err := h.casos.Ver(r.Context(), proyectoID, numero)
	if respondeErrorDePoker(w, err, "volver a mostrar la sesión de Planning Poker", mensajeNoSeCargo) {
		return
	}
	completa := h.pagina(v)
	if h.conVotosRevelados(w, r, &completa) {
		return
	}
	completa.QuienID, completa.Aviso, completa.Errores, completa.Mensaje = pagina.QuienID, pagina.Aviso, pagina.Errores, ""
	renderPoker(w, codigo, completa)
}

// mensajeNoSeCerro es el mensaje de una falla de la base al cerrar la ronda (spec de HU-19, sección 7).
const mensajeNoSeCerro = "No se pudo cerrar la ronda. Probá de nuevo en unos minutos."
