package backlog

// Estimar devuelve la historia con los story points nuevos, si el valor vale; la original no cambia. Si ya tenía una
// estimación, la reemplaza (RN5). Lo usan la página de la historia (HU-16) y, más adelante, las sesiones de Planning
// Poker para guardar el valor acordado.
//
// Spec: specs/HU-16-estimar-historia.md.
func (h Historia) Estimar(texto string) (Historia, error) {
	puntos, err := parsearStoryPoints(texto)
	if err != nil {
		return Historia{}, err
	}
	h.StoryPoints = puntos
	return h, nil
}
