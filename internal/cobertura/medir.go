// Package cobertura mide la cobertura de los tests de los paquetes que el equipo se comprometió a mantener por encima de
// un mínimo (TEC-06). Lee el perfil que escribe `go test -coverprofile` y devuelve la cobertura combinada. No usa la base
// de datos, HTTP ni el sistema de archivos: recibe el perfil como un io.Reader, así se prueba sin tocar nada.
package cobertura

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"path"
	"sort"
	"strconv"
	"strings"
)

// Paquete es la cobertura de un paquete: cuántas sentencias tiene y cuántas ejecutaron los tests.
type Paquete struct {
	Ruta       string // relativa al módulo, por ejemplo "internal/domain/backlog"
	Sentencias int
	Cubiertas  int
}

// Medicion es la cobertura de los paquetes controlados, de la que se calcula la combinada.
type Medicion struct {
	Controlados []string  // los paquetes que se pidió controlar (con sus subpaquetes)
	Paquetes    []Paquete // los que aparecen en el perfil, ordenados por ruta
}

// Sentencias es el total de sentencias de los paquetes controlados.
func (m Medicion) Sentencias() int {
	total := 0
	for _, p := range m.Paquetes {
		total += p.Sentencias
	}
	return total
}

// Cubiertas es el total de sentencias de los paquetes controlados que ejecutaron los tests.
func (m Medicion) Cubiertas() int {
	total := 0
	for _, p := range m.Paquetes {
		total += p.Cubiertas
	}
	return total
}

// Porcentaje es la cobertura combinada (RN1): sentencias cubiertas sobre sentencias totales, de 0 a 100. Un paquete con
// más sentencias pesa más. Sin sentencias es 0.
func (m Medicion) Porcentaje() float64 {
	return porcentajeDe(m.Cubiertas(), m.Sentencias())
}

func porcentajeDe(cubiertas, sentencias int) float64 {
	if sentencias == 0 {
		return 0
	}
	return float64(cubiertas) * 100 / float64(sentencias)
}

// bloque es un tramo de código del perfil. Si aparece más de una vez cuenta una sola (RN4).
type bloque struct {
	paquete    string
	sentencias int
	cubierto   bool
}

// Medir lee un perfil de cobertura y suma las sentencias de los paquetes controlados y de sus subpaquetes (RN3). Cada
// controlado es una ruta relativa al módulo, por ejemplo "internal/domain". El resto del perfil se ignora.
func Medir(perfil io.Reader, controlados []string) (Medicion, error) {
	escaner := bufio.NewScanner(perfil)
	escaner.Buffer(make([]byte, 0, 64*1024), 1024*1024)

	if !escaner.Scan() {
		if err := escaner.Err(); err != nil {
			return Medicion{}, fmt.Errorf("perfil de cobertura: %w", err)
		}
		return Medicion{}, errors.New("perfil de cobertura: está vacío")
	}
	if err := validarModo(escaner.Text()); err != nil {
		return Medicion{}, err
	}

	bloques := map[string]*bloque{}
	numero := 1
	for escaner.Scan() {
		numero++
		texto := strings.TrimSpace(escaner.Text())
		if texto == "" {
			continue
		}
		clave, archivo, sentencias, veces, ok := leerLinea(texto)
		if !ok {
			return Medicion{}, fmt.Errorf("perfil de cobertura: línea %d mal formada: %q", numero, texto)
		}
		ruta, controlado := paqueteControlado(archivo, controlados)
		if !controlado {
			continue
		}
		if b, repetido := bloques[clave]; repetido {
			b.cubierto = b.cubierto || veces > 0
			continue
		}
		bloques[clave] = &bloque{paquete: ruta, sentencias: sentencias, cubierto: veces > 0}
	}
	if err := escaner.Err(); err != nil {
		return Medicion{}, fmt.Errorf("perfil de cobertura: %w", err)
	}

	return Medicion{Controlados: append([]string(nil), controlados...), Paquetes: agrupar(bloques)}, nil
}

func validarModo(primera string) error {
	modo, ok := strings.CutPrefix(strings.TrimSpace(primera), "mode:")
	if !ok {
		return errors.New("perfil de cobertura: la primera línea tiene que empezar con \"mode:\"")
	}
	modo = strings.TrimSpace(modo)
	if modo != "set" && modo != "count" && modo != "atomic" {
		return fmt.Errorf("perfil de cobertura: modo desconocido %q", modo)
	}
	return nil
}

// leerLinea separa una línea "archivo:inicio,fin sentencias veces". La clave identifica el bloque (archivo y tramo).
func leerLinea(texto string) (clave, archivo string, sentencias, veces int, ok bool) {
	campos := strings.Fields(texto)
	if len(campos) != 3 {
		return "", "", 0, 0, false
	}
	clave = campos[0]
	dosPuntos := strings.LastIndex(clave, ":")
	if dosPuntos <= 0 || dosPuntos == len(clave)-1 {
		return "", "", 0, 0, false
	}
	sentencias, errS := strconv.Atoi(campos[1])
	veces, errV := strconv.Atoi(campos[2])
	if errS != nil || errV != nil || sentencias < 0 || veces < 0 {
		return "", "", 0, 0, false
	}
	return clave, clave[:dosPuntos], sentencias, veces, true
}

// paqueteControlado dice a qué paquete pertenece el archivo si está dentro de alguno de los controlados. Compara
// carpetas enteras: internal/domainx no es parte de internal/domain.
func paqueteControlado(archivo string, controlados []string) (string, bool) {
	carpetas := "/" + path.Dir(archivo) + "/"
	for _, c := range controlados {
		if i := strings.Index(carpetas, "/"+c+"/"); i >= 0 {
			return strings.TrimSuffix(carpetas[i+1:], "/"), true
		}
	}
	return "", false
}

func agrupar(bloques map[string]*bloque) []Paquete {
	porRuta := map[string]*Paquete{}
	for _, b := range bloques {
		p, existe := porRuta[b.paquete]
		if !existe {
			p = &Paquete{Ruta: b.paquete}
			porRuta[b.paquete] = p
		}
		p.Sentencias += b.sentencias
		if b.cubierto {
			p.Cubiertas += b.sentencias
		}
	}
	paquetes := make([]Paquete, 0, len(porRuta))
	for _, p := range porRuta {
		paquetes = append(paquetes, *p)
	}
	sort.Slice(paquetes, func(i, j int) bool { return paquetes[i].Ruta < paquetes[j].Ruta })
	return paquetes
}
