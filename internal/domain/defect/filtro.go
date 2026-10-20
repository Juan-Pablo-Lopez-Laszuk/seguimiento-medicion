package defect

// Estados son todos los estados de un defecto, en el orden del ciclo (HU-28), para los filtros (HU-29).
var Estados = []Estado{EstadoAbierto, EstadoEnProgreso, EstadoResuelto, EstadoCerrado, EstadoReabierto}

// Filtro elige defectos por severidad, estado y sprint de detección (HU-29). Un campo vacío (o 0) no filtra.
type Filtro struct {
	Severidad Severidad
	Estado    Estado
	SprintID  int64 // sprint de detección
}

// Cumple dice si el defecto cumple todo lo que el filtro eligió (CA-29.1).
func (f Filtro) Cumple(d Defecto) bool {
	return (f.Severidad == "" || d.Severidad == f.Severidad) &&
		(f.Estado == "" || d.Estado == f.Estado) &&
		(f.SprintID == 0 || d.SprintDeteccionID == f.SprintID)
}
