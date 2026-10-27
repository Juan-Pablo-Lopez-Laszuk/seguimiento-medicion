package server

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/member"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/poker"
)

// campoValor es el campo del formulario con el valor de la carta elegida (HU-18).
const campoValor = "valor"

// mensajeNoSeGuardoVoto es el mensaje de una falla de la base al votar (spec de HU-18, sección 7).
const mensajeNoSeGuardoVoto = "No se pudo guardar el voto. Probá de nuevo en unos minutos."

// votar guarda el voto de quien hace el pedido en la ronda actual (HU-18) y vuelve a la página de la sesión (303). Si
// no se puede, muestra la página con el aviso o el error en su campo y el código de la spec (sección 7).
func (h pokerHandlers) votar(w http.ResponseWriter, r *http.Request) {
	proyectoID, numero, ok := idsDeHistoria(w, r)
	if !ok {
		return
	}
	quienID, err := h.quienParticipa(r, proyectoID)
	if err == nil {
		_, err = h.casos.Votar(r.Context(), proyectoID, numero, quienID, r.PostFormValue(campoValor))
	}
	if err == nil {
		http.Redirect(w, r, fmt.Sprintf("/proyectos/%d/backlog/%d/poker?votado=1", proyectoID, numero), http.StatusSeeOther)
		return
	}

	var codigo int
	pedido := paginaPoker{QuienID: quienID, Errores: map[string]string{}}
	switch {
	case errors.Is(err, poker.ErrVotoFueraDeEscala):
		codigo, pedido.Errores[campoValor] = http.StatusUnprocessableEntity, err.Error()
	case errors.Is(err, member.ErrNoEsIntegrante):
		codigo, pedido.Aviso = http.StatusForbidden, mensajeNoEsIntegrante
	case errors.Is(err, poker.ErrSoloParticipantes) && h.cuentas == nil && quienID == 0:
		codigo, pedido.Errores[campoQuien] = http.StatusUnprocessableEntity, err.Error()
	case errors.Is(err, poker.ErrSoloParticipantes):
		codigo, pedido.Aviso = http.StatusForbidden, err.Error()
	case errors.Is(err, poker.ErrSesionNoEncontrada), errors.Is(err, poker.ErrRondaRevelada), errors.Is(err, poker.ErrSesionCerrada):
		codigo, pedido.Aviso = http.StatusConflict, err.Error()
	default:
		respondeErrorDePoker(w, err, "votar en la sesión de Planning Poker", mensajeNoSeGuardoVoto)
		return
	}
	h.mostrarConAviso(w, r, proyectoID, numero, codigo, pedido)
}

// quienVoto devuelve solo el fragmento de quién votó (bloque pokerQuienVoto), que HTMX pide cada 5 segundos (HU-18,
// T5). Nunca tiene los valores (RN5). Sin sesión abierta el fragmento queda vacío: HTMX lo saca y deja de pedirlo.
func (h pokerHandlers) quienVoto(w http.ResponseWriter, r *http.Request) {
	proyectoID, numero, ok := idsDeHistoria(w, r)
	if !ok {
		return
	}
	v, err := h.casos.Ver(r.Context(), proyectoID, numero)
	if respondeErrorDePoker(w, err, "ver quién votó", mensajeNoSeCargo) {
		return
	}
	var buf bytes.Buffer
	if err := plantillas["poker.html"].ExecuteTemplate(&buf, "pokerQuienVoto", h.pagina(v)); err != nil {
		log.Printf("render pokerQuienVoto: %v", err)
		http.Error(w, "error al mostrar la página", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = buf.WriteTo(w)
}
