# SPEC-TEC-06 · Umbral de cobertura en el CI

**Tarea:** TEC-06 (#69) · **Responsable:** Piastrellini, Mariano · **Sprint:** 3 · **Estado:** Borrador

> Especificación escrita antes del código (SDD). Usa la plantilla `specs/_plantilla.md`. TEC-06 tiene cuatro criterios de
> aceptación y se entrega en **dos PRs**: este (el umbral del CI, CA-06.2) y otro después (la matriz de trazabilidad,
> CA-06.3). Este PR dice `Refs #69`, no `Closes`: el issue se cierra con la matriz.

## 1. Objetivo

Que el CI **falle** si la cobertura de los tests de `internal/domain` y de `internal/metrics` baja del 80 %. Hoy el CI
solo *publica* la cobertura: nada impide que un PR la baje. Es una meta interna del equipo (la cátedra no fija un
porcentaje, ver la Review del Sprint 0) y figura en la Definition of Done; este PR hace que se cumpla sola.

## 2. Entradas

| Dato | Tipo | Obligatorio | Validación |
|---|---|---|---|
| Perfil de cobertura (`coverage.out`, el que escribe `go test -coverprofile`) | archivo de texto | Sí | Empieza con `mode: …` y cada línea tiene el formato `archivo:inicio,fin sentencias veces` |
| Umbral mínimo, en porcentaje | número | No (80) | Entre 0 y 100 |
| Paquetes que se controlan | rutas relativas al módulo | No (`internal/domain` e `internal/metrics`) | Cuentan también sus subpaquetes |

## 3. Salidas esperadas

- Un **informe** con una línea por paquete controlado (sentencias cubiertas / sentencias totales y porcentaje) y una última
  línea con la cobertura **combinada** y el umbral.
- Código de salida **0** si la combinada llega al umbral y **1** si no (o si algo salió mal), con el motivo en el mensaje.

```text
Cobertura de los paquetes controlados (sentencias cubiertas / totales)
  internal/domain/backlog    138/141     97,9 %
  internal/metrics            89/89     100,0 %
  Combinada                  227/230     98,7 %   (mínimo 80,0 %)
```

- En el CI, el informe queda en el resumen de la corrida (junto a la cobertura por función que ya se publica).

## 4. Reglas de negocio

- RN1: La cobertura es **sentencias cubiertas ÷ sentencias totales**, la misma medida que `go test -cover` y
  `go tool cover`. La **combinada** suma las sentencias de todos los paquetes controlados (un paquete grande pesa más que
  uno chico); no es el promedio de los porcentajes.
- RN2: El umbral se **cumple** si la combinada es mayor **o igual** que el mínimo: justo en el 80 % pasa; por debajo falla.
- RN3: Solo cuentan los paquetes controlados y sus subpaquetes (`internal/domain/backlog`, `internal/domain/sprint`…).
  El resto (`app`, `cmd`, `internal/server`, `internal/service`, `internal/store`) no cuenta para el umbral: se miden
  con los tests de integración y los escenarios BDD, pero la meta del equipo es sobre el dominio y las métricas.
- RN4: Si un mismo bloque de código aparece más de una vez en el perfil, cuenta **una sola vez** y está cubierto si alguna
  de las apariciones lo cubre (como hace `go tool cover`).
- RN5: Si el perfil **no tiene ninguna sentencia** de los paquetes controlados, el chequeo **falla** (no pasa con 0 de 0):
  lo más probable es que se haya renombrado o movido un paquete y el umbral dejaría de proteger algo sin avisar.
- RN6: Un perfil que no se puede leer (no existe, no empieza con `mode:`, una línea mal formada) hace fallar el chequeo
  con un mensaje que dice cuál es el problema; nunca pasa en silencio.

## 5. Restricciones

- Sin dependencias nuevas: es Go de la biblioteca estándar. La lógica va en `internal/cobertura` (se prueba con TDD,
  sin tocar archivos) y `cmd/cobertura` es solo la línea de comandos, como `cmd/migrar`.
- El CI lo corre **después** de los tests y **antes** del resumen de cobertura (que sigue corriendo siempre), así el
  informe se publica también cuando el umbral falla.
- El mínimo (80) y los paquetes controlados son valores por defecto de la herramienta: el CI no los repite ni los cambia.
- No cambia ningún test ni ninguna regla de la aplicación.

## 6. Casos límite

| Caso | Resultado |
|---|---|
| Combinada de 100 % | Cumple |
| Combinada justo en 80,0 % (80 de 100) | Cumple |
| Combinada de 79,0 % (79 de 100) | Falla |
| Un paquete controlado en 0 % y otro en 100 %, combinada de 90 % | Cumple (se combinan) |
| Un paquete no controlado (`cmd/server`) en 0 % | No cuenta ni aparece en el informe |
| Un subpaquete (`internal/domain/backlog`) | Cuenta como parte de `internal/domain` |
| Perfil sin ninguna sentencia de los paquetes controlados | Falla (RN5) |
| El mismo bloque dos veces, una cubierta y otra no | Una sola sentencia, cubierta (RN4) |
| Perfil vacío o sin `mode:` | Falla (RN6) |
| Línea mal formada | Falla y dice en qué línea (RN6) |

## 7. Condiciones de error

| Situación | Salida | Mensaje |
|---|---|---|
| Combinada por debajo del mínimo | código 1 | `la cobertura combinada de internal/domain e internal/metrics es 79,0 %, por debajo del mínimo de 80,0 %` |
| Sin sentencias controladas | código 1 | `el perfil no tiene sentencias de los paquetes controlados (internal/domain, internal/metrics): ¿cambió el nombre de algún paquete?` |
| Perfil ilegible o mal formado | código 1 | `perfil de cobertura: …` (con la línea) |
| El archivo del perfil no existe | código 1 | `no se pudo abrir el perfil de cobertura …` |
| Umbral fuera de 0 a 100 | código 1 | `el umbral tiene que estar entre 0 y 100` |

## 8. Criterios de aceptación (issue #69)

- CA-06.1: Tests de repositorios contra una base real con datos aislados. **Ya cumplido por TEC-03**
  (`internal/store/postgres/*_test.go`, con `TEST_DATABASE_URL`; cada test crea y borra su propio esquema) y corren en el
  CI con un PostgreSQL de servicio. Este PR solo lo deja escrito en la trazabilidad.
- CA-06.2: **El CI falla si la cobertura combinada de `internal/domain` e `internal/metrics` baja del 80 %.** (Este PR.)
- CA-06.3: Matriz requerimiento → test en `docs/trazabilidad-requisitos.md`. (El PR siguiente.)
- CA-06.4: La cobertura por paquete queda disponible para el informe. **Ya la publica el CI** (resumen + artefacto
  `cobertura`); este PR suma al resumen la tabla por paquete controlado.

## Trazabilidad

Issue: #69 · Feature: `features/tec-06-umbral-cobertura.feature` · Pasos BDD: `features/tec06_steps_test.go`

Las columnas de código y tests se completan al implementar.
