package cobertura

import (
	"errors"
	"fmt"
	"math"
	"strings"
)

const nombreCombinada = "Combinada"

// Verificar dice si la cobertura combinada llega al mínimo (RN2: justo en el mínimo cumple). Falla si queda por debajo,
// si el perfil no trae ninguna sentencia de los paquetes controlados (RN5) o si el mínimo no es un porcentaje.
func Verificar(m Medicion, minimo float64) error {
	if math.IsNaN(minimo) || minimo < 0 || minimo > 100 {
		return errors.New("el umbral tiene que estar entre 0 y 100")
	}
	if m.Sentencias() == 0 {
		return fmt.Errorf("el perfil no tiene sentencias de los paquetes controlados (%s): ¿cambió el nombre de algún paquete?",
			strings.Join(m.Controlados, ", "))
	}
	if combinada := m.Porcentaje(); combinada < minimo {
		return fmt.Errorf("la cobertura combinada de %s es %s, por debajo del mínimo de %s",
			listar(m.Controlados), porcentaje(combinada), porcentaje(minimo))
	}
	return nil
}

// Informe es el texto que ve el equipo en el CI: una línea por paquete que apareció en el perfil y la combinada al final.
func Informe(m Medicion, minimo float64) string {
	filas := make([][3]string, 0, len(m.Paquetes)+1)
	for _, p := range m.Paquetes {
		filas = append(filas, fila(p.Ruta, p.Cubiertas, p.Sentencias))
	}
	filas = append(filas, fila(nombreCombinada, m.Cubiertas(), m.Sentencias()))

	anchoNombre, anchoFraccion := 0, 0
	for _, f := range filas {
		anchoNombre = max(anchoNombre, len(f[0]))
		anchoFraccion = max(anchoFraccion, len(f[1]))
	}

	var salida strings.Builder
	salida.WriteString("Cobertura de los paquetes controlados (sentencias cubiertas / totales)\n")
	for _, f := range filas {
		fmt.Fprintf(&salida, "  %-*s  %*s  %7s", anchoNombre, f[0], anchoFraccion, f[1], f[2])
		if f[0] == nombreCombinada {
			fmt.Fprintf(&salida, "   (mínimo %s)", porcentaje(minimo))
		}
		salida.WriteString("\n")
	}
	return salida.String()
}

func fila(nombre string, cubiertas, sentencias int) [3]string {
	return [3]string{nombre, fmt.Sprintf("%d/%d", cubiertas, sentencias), porcentaje(porcentajeDe(cubiertas, sentencias))}
}

// porcentaje lo escribe como en el resto del proyecto: una decimal, con coma y con espacio antes del %.
func porcentaje(valor float64) string {
	return strings.Replace(fmt.Sprintf("%.1f %%", valor), ".", ",", 1)
}

// listar escribe "a" o "a e b" o "a, b e c".
func listar(elementos []string) string {
	if len(elementos) < 2 {
		return strings.Join(elementos, "")
	}
	return strings.Join(elementos[:len(elementos)-1], ", ") + " e " + elementos[len(elementos)-1]
}
