package memory

import (
	"context"
	"slices"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/domain/poker"
)

// Votar guarda el voto del integrante en la ronda; si ya había votado en esa ronda, lo reemplaza (HU-18, CA-18.1).
// Si la ronda no existe devuelve poker.ErrSesionNoEncontrada; si su sesión está cerrada, poker.ErrSesionCerrada; si
// ya se reveló, poker.ErrRondaRevelada (RN3). En esos casos no guarda nada.
func (r *Poker) Votar(_ context.Context, rondaID, integranteID int64, valor int) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := r.sePuedeVotarEn(rondaID); err != nil {
		return err
	}
	for i, v := range r.votos {
		if v.RondaID == rondaID && v.IntegranteID == integranteID {
			r.votos[i].Valor = valor
			return nil
		}
	}
	r.ultimoVoto++
	r.votos = append(r.votos, poker.Voto{ID: r.ultimoVoto, RondaID: rondaID, IntegranteID: integranteID, Valor: valor})
	return nil
}

// QuienesVotaron devuelve los IDs de los integrantes que votaron en la ronda, de menor a mayor. Nunca los valores
// (CA-18.3, RN5): leerlos es de HU-19, con la ronda revelada.
func (r *Poker) QuienesVotaron(_ context.Context, rondaID int64) ([]int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	ids := []int64{}
	for _, v := range r.votos {
		if v.RondaID == rondaID {
			ids = append(ids, v.IntegranteID)
		}
	}
	slices.Sort(ids)
	return ids, nil
}

// sePuedeVotarEn revisa que la ronda exista, que su sesión esté abierta y que no se haya revelado. Quien la llama
// tiene el candado.
func (r *Poker) sePuedeVotarEn(rondaID int64) error {
	i := slices.IndexFunc(r.rondas, func(ronda poker.Ronda) bool { return ronda.ID == rondaID })
	if i < 0 {
		return poker.ErrSesionNoEncontrada
	}
	ronda := r.rondas[i]
	if s := slices.IndexFunc(r.sesiones, func(s poker.Sesion) bool { return s.ID == ronda.SesionID }); r.sesiones[s].Estado == poker.EstadoCerrada {
		return poker.ErrSesionCerrada
	}
	if ronda.Revelada {
		return poker.ErrRondaRevelada
	}
	return nil
}
