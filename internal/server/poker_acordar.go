package server

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"slices"
	"strconv"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/backlog"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/member"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/poker"
)

// campoValorAcordado es el desplegable con la escala donde se elige el valor acordado (HU-22).
const campoValorAcordado = "valor"

// mensajeNoSeRegistro es el aviso de una falla de la base al registrar (spec de HU-22, sección 7).
const mensajeNoSeRegistro = "No se pudo registrar la estimación. Probá de nuevo en unos minutos."

// acordar registra el valor acordado de la sesión abierta (HU-22) y vuelve a la página de la sesión con el mensaje
// (303). Si no se puede, muestra la página con el aviso o el error del campo y el código de la spec (sección 7).
func (h pokerHandlers) acordar(w http.ResponseWriter, r *http.Request) {
	proyectoID, numero, ok := idsDeHistoria(w, r)
	if !ok {
		return
	}
	valor := r.PostFormValue(campoValorAcordado)
	quienID, err := h.quienParticipa(r, proyectoID)
	var estimada backlog.Historia
	if err == nil {
		estimada, err = h.casos.Acordar(r.Context(), proyectoID, numero, quienID, valor)
	}
	if err == nil {
		http.Redirect(w, r, fmt.Sprintf("/proyectos/%d/backlog/%d/poker?acordada=%d", proyectoID, numero, *estimada.StoryPoints), http.StatusSeeOther)
		return
	}

	var (
		codigo int
		errs   domain.ErroresValidacion
	)
	pagina := paginaPoker{QuienID: quienID, ValorElegido: valor, Errores: map[string]string{}}
	switch {
	case errors.As(err, &errs):
		codigo, pagina.Errores[campoValorAcordado] = http.StatusUnprocessableEntity, errs[backlog.CampoStoryPoints].Error()
	case errors.Is(err, poker.ErrSesionNoEncontrada):
		codigo, pagina.Aviso = http.StatusNotFound, err.Error()
	case errors.Is(err, poker.ErrAcordarSinRevelar), errors.Is(err, poker.ErrSesionCerrada),
		errors.Is(err, backlog.ErrEstimacionDeHistoriaHecha):
		codigo, pagina.Aviso = http.StatusConflict, err.Error()
	case errors.Is(err, member.ErrNoEsIntegrante):
		codigo, pagina.Aviso = http.StatusForbidden, mensajeNoEsIntegrante
	case errors.Is(err, poker.ErrSoloAgileEnablerAcuerda) && h.cuentas == nil && quienID == 0:
		codigo, pagina.Errores[campoQuien] = http.StatusUnprocessableEntity, err.Error()
	case errors.Is(err, poker.ErrSoloAgileEnablerAcuerda):
		codigo, pagina.Aviso = http.StatusForbidden, err.Error()
	default:
		respondeErrorDePoker(w, err, "registrar la estimación acordada", mensajeNoSeRegistro)
		return
	}
	v, err := h.casos.Ver(r.Context(), proyectoID, numero)
	if respondeErrorDePoker(w, err, "volver a mostrar la sesión de Planning Poker", mensajeNoSeCargo) {
		return
	}
	completa := h.pagina(v)
	completa.QuienID, completa.ValorElegido, completa.Aviso, completa.Errores = pagina.QuienID, pagina.ValorElegido, pagina.Aviso, pagina.Errores
	renderPoker(w, codigo, completa)
}

// mensajeAcordada arma el mensaje de éxito después de registrar (spec de HU-22, sección 3). Solo lo arma si el valor de
// la URL es de la escala: cualquier otra cosa se ignora, así la página nunca muestra lo que venga en la URL.
func mensajeAcordada(consulta url.Values) string {
	valor, err := strconv.Atoi(consulta.Get("acordada"))
	if err != nil || !slices.Contains(backlog.EscalaStoryPoints, valor) {
		return ""
	}
	return fmt.Sprintf("Estimación acordada: %d SP. La sesión de Planning Poker quedó cerrada.", valor)
}
