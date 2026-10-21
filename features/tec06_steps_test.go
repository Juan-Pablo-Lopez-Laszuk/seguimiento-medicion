package features

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cucumber/godog"

	"github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/internal/cobertura"
)

const moduloDelPerfil = "github.com/Juan-Pablo-Lopez-Laszuk/seguimiento-medicion/"

// tec06 guarda el estado de un escenario de TEC-06 (spec specs/TEC-06-umbral-cobertura.md): el perfil de cobertura que
// se arma con los pasos "Dado" y lo que dio verificarlo. Los pasos usan los mismos paquetes y el mismo mínimo por
// defecto que la herramienta (cmd/cobertura).
type tec06 struct {
	lineas  []string // líneas del perfil, sin el encabezado "mode:"
	crudo   string   // si no está vacío, es el perfil tal cual (un texto que no es un perfil)
	bloques int      // numera los bloques para que cada uno tenga un tramo distinto
	informe string
	err     error
}

func registrarPasosTEC06(sc *godog.ScenarioContext) {
	e := &tec06{}
	sc.Step(`^que el perfil de cobertura tiene:$`, e.elPerfilTiene)
	sc.Step(`^el perfil de cobertura tiene además:$`, e.elPerfilTiene)
	sc.Step(`^que el perfil de cobertura repite un bloque de (\d+) sentencias de "([^"]*)", una vez cubierto y otra no$`, e.elPerfilRepiteUnBloque)
	sc.Step(`^que el perfil de cobertura es un texto que no es un perfil$`, e.elPerfilNoEsUnPerfil)
	sc.Step(`^verifico el umbral de cobertura de (\d+) %$`, e.verificoElUmbral)
	sc.Step(`^el umbral se cumple$`, e.elUmbralSeCumple)
	sc.Step(`^el umbral no se cumple$`, e.elUmbralNoSeCumple)
	sc.Step(`^la cobertura combinada es "([^"]*)"$`, e.laCoberturaCombinadaEs)
	sc.Step(`^el error dice "([^"]*)"$`, e.elErrorDice)
	sc.Step(`^el error empieza con "([^"]*)"$`, e.elErrorEmpiezaCon)
	sc.Step(`^el informe tiene una línea para "([^"]*)"$`, e.elInformeTieneUnaLineaPara)
	sc.Step(`^el informe no tiene una línea para "([^"]*)"$`, e.elInformeNoTieneUnaLineaPara)
}

// bloque agrega al perfil un tramo de código de un paquete con esas sentencias, ejecutado esas veces.
func (e *tec06) bloque(paquete string, sentencias, veces int) {
	e.bloques++
	e.lineas = append(e.lineas, fmt.Sprintf("%s%s/archivo.go:%d.1,%d.2 %d %d", moduloDelPerfil, paquete, e.bloques, e.bloques, sentencias, veces))
}

// elPerfilTiene arma el perfil con una fila por paquete: las sentencias cubiertas van en un bloque que los tests
// ejecutaron y el resto en otro que no.
func (e *tec06) elPerfilTiene(tabla *godog.Table) error {
	for i, fila := range tabla.Rows[1:] { // la primera fila es el encabezado
		sentencias, errS := strconv.Atoi(fila.Cells[1].Value)
		cubiertas, errC := strconv.Atoi(fila.Cells[2].Value)
		if errS != nil || errC != nil || cubiertas > sentencias {
			return fmt.Errorf("fila %d de la tabla mal escrita: %v", i+1, fila.Cells)
		}
		paquete := fila.Cells[0].Value
		if cubiertas > 0 {
			e.bloque(paquete, cubiertas, 1)
		}
		if sentencias > cubiertas {
			e.bloque(paquete, sentencias-cubiertas, 0)
		}
	}
	return nil
}

// elPerfilRepiteUnBloque deja el mismo tramo dos veces: sin ejecutar y ejecutado (RN4).
func (e *tec06) elPerfilRepiteUnBloque(sentencias int, paquete string) error {
	e.bloques++
	tramo := fmt.Sprintf("%s%s/archivo.go:%d.1,%d.2 %d", moduloDelPerfil, paquete, e.bloques, e.bloques, sentencias)
	e.lineas = append(e.lineas, tramo+" 0", tramo+" 1")
	return nil
}

func (e *tec06) elPerfilNoEsUnPerfil() error {
	e.crudo = "esto no es un perfil de cobertura\n"
	return nil
}

func (e *tec06) verificoElUmbral(minimo int) error {
	texto := e.crudo
	if texto == "" {
		texto = "mode: atomic\n" + strings.Join(e.lineas, "\n") + "\n"
	}
	medicion, err := cobertura.Medir(strings.NewReader(texto), cobertura.PaquetesPorDefecto())
	if err != nil {
		e.err = err
		return nil
	}
	e.informe = cobertura.Informe(medicion, float64(minimo))
	e.err = cobertura.Verificar(medicion, float64(minimo))
	return nil
}

func (e *tec06) elUmbralSeCumple() error {
	if e.err != nil {
		return fmt.Errorf("se esperaba que el umbral se cumpliera y se obtuvo: %w", e.err)
	}
	return nil
}

func (e *tec06) elUmbralNoSeCumple() error {
	if e.err == nil {
		return fmt.Errorf("se esperaba que el umbral no se cumpliera y se cumplió:\n%s", e.informe)
	}
	return nil
}

// laCoberturaCombinadaEs mira el porcentaje de la línea "Combinada" del informe, sin el "(mínimo …)" del final.
func (e *tec06) laCoberturaCombinadaEs(esperada string) error {
	for _, linea := range strings.Split(e.informe, "\n") {
		if !strings.HasPrefix(strings.TrimSpace(linea), "Combinada") {
			continue
		}
		porcentaje, _, _ := strings.Cut(linea, "(mínimo")
		if !strings.HasSuffix(strings.TrimSpace(porcentaje), esperada) {
			return fmt.Errorf("la combinada tenía que ser %q y el informe dice: %q", esperada, strings.TrimSpace(linea))
		}
		return nil
	}
	return fmt.Errorf("el informe no tiene la línea de la combinada:\n%s", e.informe)
}

func (e *tec06) elErrorDice(mensaje string) error {
	if e.err == nil || e.err.Error() != mensaje {
		return fmt.Errorf("se esperaba el error %q y se obtuvo: %v", mensaje, e.err)
	}
	return nil
}

func (e *tec06) elErrorEmpiezaCon(prefijo string) error {
	if e.err == nil || !strings.HasPrefix(e.err.Error(), prefijo) {
		return fmt.Errorf("se esperaba un error que empezara con %q y se obtuvo: %v", prefijo, e.err)
	}
	return nil
}

// tieneLinea dice si el informe tiene una línea que empiece con ese paquete.
func (e *tec06) tieneLinea(paquete string) bool {
	for _, linea := range strings.Split(e.informe, "\n") {
		if campos := strings.Fields(linea); len(campos) > 0 && campos[0] == paquete {
			return true
		}
	}
	return false
}

func (e *tec06) elInformeTieneUnaLineaPara(paquete string) error {
	if !e.tieneLinea(paquete) {
		return fmt.Errorf("el informe tenía que tener una línea para %q:\n%s", paquete, e.informe)
	}
	return nil
}

func (e *tec06) elInformeNoTieneUnaLineaPara(paquete string) error {
	if e.tieneLinea(paquete) {
		return fmt.Errorf("el informe no tenía que tener una línea para %q:\n%s", paquete, e.informe)
	}
	return nil
}
