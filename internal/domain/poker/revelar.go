package poker

import (
	"errors"
	"sort"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/member"
)

// Errores de revelar una ronda (sección 7 de la spec de HU-19).
var (
	ErrRondaSinVotos   = errors.New("la ronda no tiene votos: no se puede revelar")
	ErrRondaYaRevelada = errors.New("la ronda ya está revelada")
	ErrRondaSinRevelar = errors.New("los votos se ven cuando la ronda se revela")
)

// VotoRevelado es una fila de los votos de una ronda revelada: el integrante y su voto; Voto es false si no votó
// (CA-19.2).
type VotoRevelado struct {
	Integrante string
	Valor      int
	Voto       bool
}

// TodosVotaron dice si votaron todos los integrantes activos (HU-19 RN1). Sin integrantes activos es false.
func TodosVotaron(votos []Voto, integrantes []member.Integrante) bool {
	votaron := map[int64]bool{}
	for _, v := range votos {
		votaron[v.IntegranteID] = true
	}
	activos := 0
	for _, i := range integrantes {
		if !i.Activo {
			continue
		}
		activos++
		if !votaron[i.ID] {
			return false
		}
	}
	return activos > 0
}

// RevelarSiCorresponde devuelve la ronda revelada si votaron todos los integrantes activos (CA-19.1, RN1); si no, la
// devuelve como estaba.
func RevelarSiCorresponde(r Ronda, votos []Voto, integrantes []member.Integrante) Ronda {
	if len(votos) > 0 && TodosVotaron(votos, integrantes) {
		r.Revelada = true
	}
	return r
}

// Cerrar revela la ronda con los votos que haya, si la cierra el Agile Enabler (CA-19.1, RN2 y RN4, con PuedeConducir de HU-17). Una ronda sin
// votos (CA-19.3) o ya revelada (RN5) no se cierra.
func Cerrar(r Ronda, votos []Voto, quien member.Integrante) (Ronda, error) {
	if err := PuedeConducir(quien); err != nil {
		return r, err
	}
	switch {
	case r.Revelada:
		return r, ErrRondaYaRevelada
	case len(votos) == 0:
		return r, ErrRondaSinVotos
	}
	r.Revelada = true
	return r, nil
}

// VotosRevelados devuelve cada integrante activo con su voto, y también quien votó aunque ya no esté activo, ordenados
// por nombre (CA-19.2). Si la ronda no está revelada no devuelve ningún valor (RN6).
func VotosRevelados(r Ronda, votos []Voto, integrantes []member.Integrante) ([]VotoRevelado, error) {
	if !r.Revelada {
		return nil, ErrRondaSinRevelar
	}
	valor := map[int64]int{}
	for _, v := range votos {
		valor[v.IntegranteID] = v.Valor
	}
	var filas []VotoRevelado
	for _, i := range integrantes {
		v, voto := valor[i.ID]
		if i.Activo || voto {
			filas = append(filas, VotoRevelado{Integrante: i.Nombre, Valor: v, Voto: voto})
		}
	}
	sort.Slice(filas, func(a, b int) bool { return filas[a].Integrante < filas[b].Integrante })
	return filas, nil
}
