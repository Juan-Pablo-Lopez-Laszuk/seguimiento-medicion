# SPEC-33 · Horas estimadas, reales y desviación

**Historia:** HU-33 (#11) · **Responsable:** Bravo, Carolina · **Sprint:** 3 · **Estado:** Implementada

> Especificación escrita antes del código (SDD). Usa las métricas M4, M5 y M6 de `docs/metricas.md`.

## 1. Objetivo

Comparar las horas estimadas con las reales para ver cuánto se equivocó la estimación, por historia, por sprint y
por proyecto. La usan el panel (HU-35) y el reporte de sprint (HU-37).

## 2. Entradas

| Dato | De dónde sale |
|---|---|
| Proyecto | id (viene en la URL); tiene que existir |
| Horas estimadas de una historia (M4) | Suma de las horas estimadas de sus tareas (HU-25). Una tarea sin estimar vale 0 |
| Horas reales de una historia (M5) | Suma de los registros de esfuerzo de la historia (HU-24) |
| Sprint de la historia | El sprint al que está asignada (HU-12); una historia sin sprint solo cuenta para el proyecto |

## 3. Salidas esperadas

- **Página de horas** (`/proyectos/{id}/horas`), con enlace desde la vista del proyecto:
  - El total del proyecto: horas estimadas, reales, desviación en horas y en %.
  - Una fila por sprint con lo mismo.
  - Una fila por historia con lo mismo.
- La desviación en horas se muestra con signo (`+2 h`, `-2 h`, `0 h`) y el % con un decimal (`+20,0 %`).

## 4. Reglas de negocio

- RN1: Desviación (h) = reales − estimadas. Desviación (%) = (reales − estimadas) / estimadas × 100 (CA-33.1).
- RN2: Positivo = se tardó más de lo estimado; negativo = se tardó menos.
- RN3: Si las estimadas son 0, el % dice **"no aplicable"**; no es un error (CA-33.2).
- RN4: El sprint y el proyecto **suman** las horas de sus historias y recién ahí calculan el %; no se promedian los %
  de cada historia (CA-33.3).

## 5. Restricciones

- Cálculo puro en `internal/metrics/deviation.go` (`CalcularDesviacion`, `SumarDesviaciones`), como el resto de las
  métricas. El caso de uso `service.Horas` junta historias, tareas y registros. Solo lee: no cambia nada.
- Las horas cargadas en la historia misma (`historias.horas_estimadas`) todavía no se guardan desde la app (HU-23
  dejó la regla), así que una historia sin tareas tiene 0 h estimadas.
- Sin migración nueva.

## 6. Casos límite

| Caso | Resultado |
|---|---|
| Estimadas 10 h, reales 12 h | +2 h y +20,0 % |
| Estimadas 10 h, reales 8 h | -2 h y -20,0 % |
| Estimadas 0 h, reales 3 h | +3 h y "no aplicable" |
| Sin tareas ni registros | 0 h y "no aplicable" |
| Proyecto sin historias | Totales en 0 h y "no aplicable" |

## 7. Condiciones de error

| Condición | Error | Mensaje / respuesta |
|---|---|---|
| Proyecto inexistente | `project.ErrNoEncontrado` | Página 404 |

## 8. Criterios de aceptación

- CA-33.1: Desviación = reales − estimadas, también en %.
- CA-33.2: Si las estimadas son 0, el % dice "no aplicable".
- CA-33.3: Se calcula por historia, sprint y proyecto.

## Trazabilidad

Issue: #11 · Métrica: M6 · Feature: `features/hu-33-desviacion.feature` · Pasos BDD: `features/hu33_steps_test.go`

| Capa | Código | Tests |
|---|---|---|
| Métrica | `internal/metrics/deviation.go` (`CalcularDesviacion`, `SumarDesviaciones`, `HorasTexto`, `PorcentajeTexto`) | `internal/metrics/deviation_test.go` |
| Caso de uso | `internal/service/horas.go` (`Horas.Ver`) | `internal/service/horas_test.go` |
| Pantalla | `internal/server/horas.go`, `web/templates/paginas/horas.html`, botón "Horas" en `proyecto.html` | `internal/server/horas_test.go` |
| Aplicación | `app/app.go`, ruta en `internal/server/router.go` | `app/app_test.go` (`TestNewHandlerConBase_ConURLLasHorasSeCalculanDesdeLaBase`) |

| Criterio | Escenarios BDD | Tests |
|---|---|---|
| CA-33.1 Desviación en h y en % | Una historia que llevó más horas · Una historia que llevó menos horas | `TestCalcularDesviacion`, `TestPaginaDeHoras_MuestraLaDesviacion` |
| CA-33.2 Sin estimadas, "no aplicable" | Una historia sin horas estimadas | `TestCalcularDesviacion`, `TestPaginaDeHoras_SinEstimadas_NoAplica` |
| CA-33.3 Por historia, sprint y proyecto | Desviación por sprint y por proyecto | `TestSumarDesviaciones`, `TestVerHoras_DesviacionPorHistoriaSprintYProyecto` |
