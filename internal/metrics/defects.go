package metrics

// Defecto es lo que necesitan las métricas M8 y M9 de un defecto (HU-34).
type Defecto struct {
	Severidad        string
	SprintDeteccion  int64
	SprintResolucion int64 // 0 si no tiene (abierto, en progreso o reabierto)
	Resuelto         bool  // true si está Resuelto o Cerrado
}

// ConteoSprint son los defectos detectados y resueltos en un sprint.
type ConteoSprint struct {
	Detectados int
	Resueltos  int
}

// ConteoDefectos es el resultado de M8 y M9: por sprint (por ID) y el total del proyecto.
type ConteoDefectos struct {
	PorSprint  map[int64]ConteoSprint
	Detectados int
	Resueltos  int
	Abiertos   int // Detectados − Resueltos
}

// ContarDefectos cuenta los detectados por su sprint de detección y los resueltos por su sprint de resolución
// (CA-34.1). Con severidad distinta de "" cuenta solo los de esa severidad (CA-34.2). Un resuelto necesita sprint de
// resolución: un reabierto lo pierde y deja de contar.
func ContarDefectos(defectos []Defecto, severidad string) ConteoDefectos {
	c := ConteoDefectos{PorSprint: map[int64]ConteoSprint{}}
	for _, d := range defectos {
		if severidad != "" && d.Severidad != severidad {
			continue
		}
		det := c.PorSprint[d.SprintDeteccion]
		det.Detectados++
		c.PorSprint[d.SprintDeteccion] = det
		c.Detectados++
		if d.Resuelto && d.SprintResolucion != 0 {
			res := c.PorSprint[d.SprintResolucion]
			res.Resueltos++
			c.PorSprint[d.SprintResolucion] = res
			c.Resueltos++
		}
	}
	c.Abiertos = c.Detectados - c.Resueltos
	return c
}
