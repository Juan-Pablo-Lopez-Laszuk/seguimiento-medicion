# SPEC-32 · Porcentaje de historias completadas

**Historia:** HU-32 (#10) · **Responsable:** Bravo, Carolina · **Sprint:** 1 · **Estado:** Implementada

> Especificación escrita antes del código (SDD). Usa la plantilla `specs/_plantilla.md` y la fórmula M7 de `docs/metricas.md`.

## 1. Objetivo

Que el equipo vea qué parte de las historias que se comprometió a hacer en un sprint terminó de verdad, contando historias y no story points. Lo usa el Agile Enabler al cerrar el sprint y se muestra en el dashboard.

## 2. Entradas

| Dato | Tipo | Obligatorio | Validación |
|---|---|---|---|
| Historias del sprint | lista de historias | Sí (puede estar vacía) | — |
| Estado de cada historia | `Pendiente`, `En sprint`, `En progreso` o `Hecho` | Sí | Solo `Hecho` cuenta como completada |
| Resultados por sprint (para el proyecto) | lista de resultados | Sí (puede estar vacía) | — |

## 3. Salidas esperadas

- **Historias Hechas** y **historias planificadas** (los dos números del cálculo).
- **Porcentaje de completadas**, con decimales internamente.
- **Texto para pantalla**, con 1 decimal y coma decimal: "75,0 %".

## 4. Reglas de negocio

- RN1: % completadas = historias Hechas ÷ historias planificadas × 100 (M7).
- RN2: Se cuentan historias, no story points: una historia de 13 SP y una de 1 SP valen lo mismo.
- RN3: Una historia sin estimar también cuenta como planificada.
- RN4: El porcentaje del proyecto se calcula con el total de historias de todos sus sprints, **no** promediando los porcentajes de cada sprint (3/4 y 3/3 dan 6/7 = 85,7 %, no 87,5 %).
- RN5: Para un sprint cerrado se usan las historias de la foto del cierre (`sprint_historias`, ver `docs/modelo-datos.md`).

## 5. Restricciones

- Cálculo puro en `internal/metrics` (sin base de datos ni HTTP), así se prueba con TDD sin depender de nadie.
- Reutiliza `metrics.Historia` de HU-30.
- Internamente `float64`; el redondeo a 1 decimal es solo para mostrar (RG-2).

## 6. Casos límite

| Caso | Resultado |
|---|---|
| Sprint sin historias | 0,0 %, sin error |
| Todas las historias Hechas | 100,0 % |
| Ninguna historia Hecha | 0,0 % |
| Proyecto sin sprints | 0,0 % |
| Resultado periódico (2 de 3) | 66,7 % |

## 7. Condiciones de error

No hay datos de entrada que puedan ser inválidos: se cuentan historias. El único riesgo es dividir por cero cuando no hay historias, y está controlado (CA-32.2): devuelve 0 %, sin error.

## 8. Criterios de aceptación

- CA-32.1: % completadas = historias Hechas ÷ historias planificadas × 100, con 1 decimal.
- CA-32.2: Un sprint sin historias da 0 % y no es un error.
- CA-32.3: Se calcula por sprint y por proyecto.

## Trazabilidad

Issue: #10 · Métrica: M7 en `docs/metricas.md` · Feature: `features/hu-32-porcentaje.feature` · Pasos BDD: `features/hu32_steps_test.go` · Tests: `internal/metrics/completion_test.go` · Código: `internal/metrics/completion.go` (`CalcularPorcentajeCompletadas`, `SumarPorcentajes`, `Texto`)

| Criterio | Escenarios BDD | Tests |
|---|---|---|
| CA-32.1 | Porcentaje de un sprint · Sprint con todas o ninguna historia hecha | `TestCalcularPorcentajeCompletadas_HechasSobrePlanificadas`, `TestPorcentaje_TextoConUnDecimal` |
| CA-32.2 | Sprint sin historias | `TestCalcularPorcentajeCompletadas_SprintSinHistoriasDaCero` |
| CA-32.3 | Porcentaje del proyecto con varios sprints | `TestSumarPorcentajes_TotalDelProyecto`, `TestSumarPorcentajes_ProyectoSinHistorias` |
