# SPEC-30 · Story points planificados y completados

**Historia:** HU-30 (#8) · **Responsable:** Bravo, Carolina · **Sprint:** 1 · **Estado:** Borrador

> Especificación escrita antes del código (SDD). Usa la plantilla `specs/_plantilla.md` y las fórmulas M1 y M2 de `docs/metricas.md`.

## 1. Objetivo

Que el equipo vea cuántos story points se comprometió a hacer en un sprint (planificados) y cuántos terminó de verdad (completados), para medir cuánto cumplió. Lo usa el Agile Enabler al cerrar el sprint y se muestra en el dashboard.

## 2. Entradas

| Dato | Tipo | Obligatorio | Validación |
|---|---|---|---|
| Historias del sprint | lista de historias | Sí (puede estar vacía) | — |
| Story points de cada historia | número | No (puede estar sin estimar) | SP ≥ 0 |
| Estado de cada historia | `Pendiente`, `En sprint`, `En progreso` o `Hecho` | Sí | Solo `Hecho` cuenta como completada |
| Resultados por sprint (para el total del proyecto) | lista de resultados | Sí (puede estar vacía) | — |

## 3. Salidas esperadas

- **SP planificados:** suma de los story points de todas las historias del sprint.
- **SP completados:** suma de los story points de las historias en estado `Hecho`.
- **Historias sin estimar:** cuántas son, y la advertencia "hay N historias sin estimar" (en singular: "hay 1 historia sin estimar").
- **Total del proyecto:** los mismos tres datos, sumando todos los sprints.

## 4. Reglas de negocio

- RN1: SP planificados = Σ SP de las historias asignadas al sprint (M1).
- RN2: SP completados = Σ SP de las historias del sprint en estado `Hecho` (M2).
- RN3: Una historia sin estimar suma 0 en planificados y en completados, y se cuenta para la advertencia.
- RN4: El total del proyecto es la suma de los resultados de todos sus sprints.
- RN5: Una historia estimada en 0 SP está estimada: suma 0 y no genera advertencia.

## 5. Restricciones

- Cálculo puro en `internal/metrics` (sin base de datos ni HTTP), así se prueba con TDD sin depender de nadie.
- Internamente `float64`; el redondeo a 1 decimal es solo para mostrar.
- Reutiliza el error `ErrSPNegativos` que ya usa la velocidad (HU-31).

## 6. Casos límite

| Caso | Resultado |
|---|---|
| Sprint sin historias | planificados 0, completados 0, sin advertencia |
| Ninguna historia Hecha | completados 0 |
| Todas las historias Hechas | completados = planificados |
| Historia de 0 SP | suma 0 y no se avisa |
| Todas las historias sin estimar | 0 y 0, con advertencia |
| Proyecto sin sprints | todo en 0, sin advertencia |

## 7. Condiciones de error

| Situación | Error | Mensaje al usuario |
|---|---|---|
| Alguna historia estimada con SP negativos | `ErrSPNegativos` | "los story points no pueden ser negativos" |

## 8. Criterios de aceptación

- CA-30.1: Planificados = suma de SP del sprint; completados = suma de SP de las historias Hechas.
- CA-30.2: Las historias sin estimar suman 0 y se avisa.
- CA-30.3: El total del proyecto es la suma de los sprints.

## Trazabilidad

Issue: #8 · Métricas: M1 y M2 en `docs/metricas.md` · Feature: `features/hu-30-sp.feature` · Pasos BDD: `features/hu30_steps_test.go` · Tests: `internal/metrics/storypoints_test.go` · Código: `internal/metrics/storypoints.go` (`CalcularStoryPoints`, `SumarStoryPoints`)

| Criterio | Escenarios BDD | Tests |
|---|---|---|
| CA-30.1 | Planificados y completados de un sprint · Sprint sin historias | `TestCalcularStoryPoints_PlanificadosYCompletados` |
| CA-30.2 | Las historias sin estimar suman 0 y se avisa | `TestCalcularStoryPoints_HistoriasSinEstimarSumanCeroYSeAvisa` |
| CA-30.3 | El total del proyecto es la suma de sus sprints | `TestSumarStoryPoints_TotalDelProyecto`, `TestSumarStoryPoints_ProyectoSinSprints` |
| Error | Story points negativos | `TestCalcularStoryPoints_SPNegativosDevuelveError` |
