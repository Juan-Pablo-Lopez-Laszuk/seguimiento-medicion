package server

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/effort"
	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/service"
)

// registrosEsfuerzo agrupa los handlers para consultar y corregir los registros de horas de un proyecto (HU-26).
type registrosEsfuerzo struct {
	casos *service.RegistrosEsfuerzo
}

// paginaRegistros son los datos de registros_esfuerzo.html: la lista filtrada y el mensaje de una corrección.
type paginaRegistros struct {
	service.VistaRegistros
	Mensaje string
}

// paginaCorreccion son los datos de corregir_registro.html: el registro, lo que se escribió y sus errores.
type paginaCorreccion struct {
	service.VistaCorreccion
	QuienID int64
	Datos   effort.DatosRegistro
	Errores map[string]string
}

// consultar muestra los registros que cumplen los filtros de la URL (?integrante, ?historia y ?sprint). Un valor que
// no es un número es como no haber elegido nada (RN3).
func (h registrosEsfuerzo) consultar(w http.ResponseWriter, r *http.Request) {
	proyectoID, ok := idDeProyecto(w, r)
	if !ok {
		return
	}
	q := r.URL.Query()
	integrante, _ := strconv.ParseInt(q.Get("integrante"), 10, 64)
	historia, _ := strconv.Atoi(q.Get("historia"))
	numeroSprint, _ := strconv.Atoi(q.Get("sprint"))
	vista, err := h.casos.Consultar(r.Context(), proyectoID, service.FiltroRegistros{IntegranteID: integrante, Historia: historia, Sprint: numeroSprint})
	if respondeErrorDeRegistros(w, err, "consultar los registros de horas") {
		return
	}
	pagina := paginaRegistros{VistaRegistros: vista}
	if q.Get("corregido") == "1" {
		pagina.Mensaje = "Registro corregido."
	}
	render(w, http.StatusOK, "registros_esfuerzo.html", Pagina{Titulo: "Registros de horas", Datos: pagina})
}

// verCorreccion muestra el formulario con los datos del registro.
func (h registrosEsfuerzo) verCorreccion(w http.ResponseWriter, r *http.Request) {
	proyectoID, registroID, ok := idsDeRegistro(w, r)
	if !ok {
		return
	}
	vista, err := h.casos.VerCorreccion(r.Context(), proyectoID, registroID)
	if respondeErrorDeRegistros(w, err, "ver la corrección del registro") {
		return
	}
	reg := vista.Registro
	datos := effort.DatosRegistro{TareaID: reg.TareaID, Fecha: fechaParaInput(reg.Fecha), Actividad: reg.Actividad, Horas: effort.FormatearHoras(reg.Horas)}
	renderCorreccion(w, http.StatusOK, paginaCorreccion{VistaCorreccion: vista, Datos: datos})
}

// corregir recibe el formulario. Si la corrección es válida vuelve a la lista (303); si no, muestra el formulario con
// cada error debajo de su campo y lo que se escribió.
func (h registrosEsfuerzo) corregir(w http.ResponseWriter, r *http.Request) {
	proyectoID, registroID, ok := idsDeRegistro(w, r)
	if !ok {
		return
	}
	// Quién corrige o una tarea que no son un número son como no haber elegido ninguno.
	quien, _ := strconv.ParseInt(r.PostFormValue("quien"), 10, 64)
	tarea, _ := strconv.ParseInt(r.PostFormValue("tarea"), 10, 64)
	d := effort.DatosRegistro{TareaID: tarea, Fecha: r.PostFormValue("fecha"), Actividad: r.PostFormValue("actividad"), Horas: r.PostFormValue("horas")}
	_, err := h.casos.Corregir(r.Context(), proyectoID, registroID, quien, d)

	var errs domain.ErroresValidacion
	switch {
	case err == nil:
		http.Redirect(w, r, fmt.Sprintf("/proyectos/%d/esfuerzo?corregido=1", proyectoID), http.StatusSeeOther)
	case errors.As(err, &errs):
		vista, err := h.casos.VerCorreccion(r.Context(), proyectoID, registroID)
		if respondeErrorDeRegistros(w, err, "volver a mostrar la corrección") {
			return
		}
		pagina := paginaCorreccion{VistaCorreccion: vista, QuienID: quien, Datos: d, Errores: make(map[string]string, len(errs))}
		for campo, e := range errs {
			pagina.Errores[campo] = e.Error()
		}
		renderCorreccion(w, http.StatusUnprocessableEntity, pagina)
	default:
		respondeErrorDeRegistros(w, err, "corregir el registro")
	}
}

func renderCorreccion(w http.ResponseWriter, codigo int, pagina paginaCorreccion) {
	render(w, codigo, "corregir_registro.html", Pagina{Titulo: "Corregir registro", Datos: pagina})
}

// idsDeRegistro lee el {id} del proyecto y el {registro} de la URL. Si el registro no es un número responde 404, igual
// que si no existiera.
func idsDeRegistro(w http.ResponseWriter, r *http.Request) (proyectoID, registroID int64, ok bool) {
	proyectoID, ok = idDeProyecto(w, r)
	if !ok {
		return 0, 0, false
	}
	registroID, err := strconv.ParseInt(chi.URLParam(r, "registro"), 10, 64)
	if err != nil {
		registroNoEncontrado(w)
		return 0, 0, false
	}
	return proyectoID, registroID, true
}

func registroNoEncontrado(w http.ResponseWriter) {
	http.Error(w, "No encontramos ese registro de horas.", http.StatusNotFound)
}

// respondeErrorDeRegistros responde 404 si el registro no existe y, si no, como los demás errores de esfuerzo.
func respondeErrorDeRegistros(w http.ResponseWriter, err error, accion string) bool {
	if errors.Is(err, effort.ErrRegistroNoEncontrado) {
		registroNoEncontrado(w)
		return true
	}
	return respondeErrorDeTareas(w, err, accion)
}
