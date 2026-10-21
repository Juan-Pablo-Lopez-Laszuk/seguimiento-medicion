package cobertura

// MinimoPorDefecto es la cobertura combinada mínima que se exige, en porcentaje. Es una meta interna del equipo (la
// cátedra no fija un porcentaje) y figura en la Definition of Done.
const MinimoPorDefecto = 80.0

// PaquetesPorDefecto son los paquetes sobre los que se exige el mínimo: el dominio (con todos sus subpaquetes) y las
// métricas. Devuelve una copia nueva cada vez.
func PaquetesPorDefecto() []string {
	return []string{"internal/domain", "internal/metrics"}
}
