# SPEC-37 · Reporte de sprint

**Historia:** HU-37 (#48) · **Responsable:** López, Juan Pablo · **Sprint:** 3 · **Estado:** Implementada

> Especificación escrita antes del código (SDD). Junta en una página las métricas de `docs/metricas.md` que ya están
> hechas: SP planificados y completados (HU-30), cumplimiento y foto del cierre (HU-15), % de historias completadas
> (HU-32), velocidad (HU-31), horas estimadas, reales y desviación (HU-23, HU-24 y HU-33) y defectos detectados y
> resueltos (HU-34). No calcula nada nuevo: arma el reporte con esas funciones (CA-37.2). Es la base del reporte de
> proyecto (HU-38) y del PDF (HU-39).
> Revisada por Juan Pablo el 20/10: pantalla nueva con botón Imprimir, foto del cierre en los sprints cerrados, reporte
> parcial en los abiertos y velocidad con los sprints cerrados hasta ese.

## 1. Objetivo

Que el Agile Enabler tenga, en una sola página, todo lo que se presenta en la Sprint Review de un sprint: qué se
planificó, qué se completó, cuánto se estimó y cuánto se trabajó, cómo dieron las métricas y qué defectos aparecieron
y se resolvieron. La página se puede imprimir y se lee bien en el celular.

## 2. Entradas

| Dato | Tipo | Obligatorio | Validación |
|---|---|---|---|
| Proyecto | id del proyecto (viene en la URL) | Sí | Tiene que existir |
| Sprint | número del sprint (viene en la URL) | Sí | Tiene que existir en el proyecto |

No hay datos que cargue el usuario: es una consulta. La página es `/proyectos/{id}/sprints/{numero}/reporte` y se
llega con el botón **Reporte del sprint** de la página del sprint (HU-12).

## 3. Salidas esperadas

Con el ejemplo de la sección 6 (Sprint 1 cerrado), el reporte muestra, en este orden:

1. **Encabezado:** "Reporte del Sprint 1 · Software Metrics", el Sprint Goal, las fechas (12/10/2026 al 18/10/2026)
   y el estado (**Cerrado**). Si el sprint no está cerrado, un aviso: "El sprint no está cerrado: el reporte muestra
   cómo va hasta hoy." (RN2).
2. **Métricas** (CA-37.1), en tarjetas:

   | Métrica | Valor | De dónde sale |
   |---|---|---|
   | SP planificados y completados | 5 de 13 SP | M1 y M2 (HU-30), de la foto del cierre |
   | Cumplimiento | 38,5 % | HU-15 |
   | Historias completadas | 1 de 3 (33,3 %) | M7 (HU-32) |
   | Velocidad del equipo | 5,0 SP por sprint (1 sprint cerrado) | M3 (HU-31), con los sprints cerrados hasta este |
   | Horas estimadas y reales | 10 h estimadas · 11 h reales | M4 (HU-23) y M5 (HU-24) |
   | Desviación | +1 h (+10,0 %) | M6 (HU-33) |
   | Defectos | 2 detectados · 1 resuelto | M8 y M9 (HU-34) |

3. **Historias del sprint** (CA-37.1), una tabla ordenada por número:

   | Historia | Título | SP | Resultado | Horas estimadas | Horas reales | Desviación |
   |---|---|---|---|---|---|---|
   | HU-1 | Login | 5 | Completada | 6 h | 8 h | +2 h (+33,3 %) |
   | HU-2 | Registro | 3 | No completada | 4 h | 3 h | -1 h (-25,0 %) |
   | HU-3 | Perfil | 5 | No completada | 0 h | 0 h | 0 h (no aplicable) |

4. **Defectos del sprint** (CA-37.1): los detectados o resueltos en este sprint, ordenados por ID:

   | Defecto | Severidad | Estado | Historia | Detectado en | Resuelto en |
   |---|---|---|---|---|---|
   | #1 El login no valida el correo | Crítica | Resuelto | HU-1 | Sprint 1 | Sprint 1 |
   | #2 El registro tarda mucho | Media | Abierto | HU-2 | Sprint 1 | — |

   Sin defectos: "No hubo defectos en este sprint."

5. Un botón **Imprimir** (la impresión oculta la barra de navegación y los botones).

## 4. Reglas de negocio

- RN1: **Un sprint cerrado usa la foto del cierre** (`docs/metricas.md` RG-6): las historias del reporte son las de la
  foto, también las que volvieron al backlog, con los SP y el resultado que tenían al cerrar. Los SP planificados y
  completados son los guardados al cerrar (HU-14). Es el mismo detalle que muestra la página del sprint (HU-15).
- RN2: **Un sprint abierto** (activo o planificado) usa sus historias de hoy, con sus SP y estado actuales, y avisa que
  el reporte es parcial. Los SP se calculan con `CalcularStoryPoints` (HU-30); si hay historias sin estimar, se muestra
  la advertencia de HU-30 ("hay 1 historia sin estimar").
- RN3: **Historias completadas** = historias completadas ÷ historias del reporte, con la regla de HU-32: sin historias
  da 0,0 %.
- RN4: **Velocidad** = la de HU-31 (últimos 3 sprints cerrados) contando solo los sprints cerrados **con número menor
  o igual** al del reporte. Así el reporte de un sprint viejo muestra la velocidad que había en ese momento. Sin sprints
  cerrados hasta ese: "Aún no hay sprints cerrados".
- RN5: **Horas** de cada historia: las estimadas son la suma de sus tareas (M4) y las reales la suma de sus registros
  (M5), igual que la página de horas de HU-33. Las del sprint son la suma de sus historias y la desviación del sprint se
  calcula sobre esa suma, no promediando (HU-33 CA-33.3). Sin horas estimadas, el % es "no aplicable".
- RN6: **Defectos:** detectados = los que tienen este sprint como sprint de detección (M8); resueltos = los que están
  Resueltos o Cerrados y tienen este sprint como sprint de resolución (M9). La tabla lista los que cumplen alguna de las
  dos. Un defecto reabierto no cuenta como resuelto (HU-28).
- RN7: Es solo lectura: consultar el reporte no cambia nada.

## 5. Restricciones

- **Sin cálculos nuevos (CA-37.2):** todas las cuentas salen de `internal/metrics` (`CalcularStoryPoints`,
  `StoryPoints.CumplimientoTexto`, `CalcularPorcentajeCompletadas`, `CalcularVelocidad`, `CalcularDesviacion`,
  `SumarDesviaciones` y `ContarDefectos`). El caso de uso solo junta los datos y llama a esas funciones.
- **Sin migración ni repositorios nuevos:** se usan los listados que ya existen (sprints, historias, foto del cierre,
  tareas y registros por historia, defectos del proyecto). La tarea T2 queda en verificar que alcanzan.
- **Caso de uso propio:** `service.Reportes` con `Sprint(ctx, proyectoID, numero) (ReporteSprint, error)`. Devuelve
  un struct con todo el reporte, así HU-38 arma el del proyecto con un `ReporteSprint` por sprint y HU-39 lo pasa a PDF
  sin volver a calcular.
- Las horas de una historia se calculan con la misma función que la página de horas (HU-33), así el reporte y esa
  página no pueden dar distinto.
- **Pantalla nueva** con el layout de siempre: tablas dentro de `table-responsive` para el celular y `@media print`
  para imprimir.
- **Depende de HU-15 y HU-34:** usa la foto y el cumplimiento de HU-15 (#413) y el resumen de defectos de HU-34
  (#119, `VistaDefectos.Resumen` con `ContarDefectos`). Las dos ya están en `main`.

## 6. Casos límite

Proyecto "Software Metrics" con el Sprint 1 (12/10 al 18/10) cerrado con HU-1 Login (5 SP, completada), HU-2 Registro
(3 SP, no) y HU-3 Perfil (5 SP, no), y el Sprint 2 (19/10 al 25/10) abierto. Horas: HU-1 con 6 h estimadas en tareas y
8 h registradas; HU-2 con 4 h estimadas y 3 h registradas; HU-3 sin tareas ni registros. Defectos: #1 Crítica de HU-1,
detectado y resuelto en el Sprint 1; #2 Media de HU-2, detectado en el Sprint 1 y abierto; #3 Baja, detectado en el
Sprint 2.

| Caso | Resultado |
|---|---|
| Reporte del Sprint 1 | Lo de la sección 3 |
| HU-2 se re-estima en 8 y entra al Sprint 2 | El reporte del Sprint 1 sigue con HU-2 en 3 SP y No completada (RN1) |
| Reporte del Sprint 2 (abierto) | Aviso de reporte parcial; sus historias de hoy; #3 en defectos; velocidad 5,0 (RN4) |
| Sprint sin historias | 0 de 0 SP, cumplimiento —, 0 de 0 historias (0,0 %), 0 h y "no aplicable" |
| Sprint sin defectos | 0 detectados · 0 resueltos y "No hubo defectos en este sprint." |
| #2 se resuelve en el Sprint 2 | Sprint 1: sigue con 2 detectados y 1 resuelto; Sprint 2: suma 1 resuelto y lista #2 (RN6) |
| #1 se reabre | Sprint 1: 2 detectados y 0 resueltos (RN6) |
| Primer sprint, todavía abierto | Velocidad: "Aún no hay sprints cerrados" (RN4) |
| Historia sin estimar en un sprint abierto | Suma 0 SP y aparece la advertencia de HU-30 (RN2) |

## 7. Condiciones de error

| Situación | Error | Mensaje al usuario |
|---|---|---|
| El proyecto no existe | 404 | "No encontramos ese proyecto." (el de HU-11) |
| El sprint no existe en el proyecto (CA-37.3) | 404 | "No encontramos ese sprint." (el de HU-12) |
| Falla al leer la base | 500 | — |

## 8. Criterios de aceptación

- CA-37.1 Incluye historias planificadas y completadas, estimaciones, esfuerzo, métricas y defectos del sprint.
- CA-37.2 Reutiliza los cálculos de `internal/metrics`.
- CA-37.3 Sprint inexistente → error 404.

## Trazabilidad

Issue: #48 · Feature: `features/hu-37-reporte-sprint.feature` · Pasos BDD: `features/hu37_steps_test.go`

| Capa | Código | Tests |
|---|---|---|
| Métricas | sin cambios: `CalcularStoryPoints`, `StoryPoints.CumplimientoTexto`, `CalcularPorcentajeCompletadas`, `CalcularVelocidad`, `CalcularDesviacion`, `SumarDesviaciones` y `ContarDefectos` (CA-37.2) | los de cada métrica |
| Caso de uso | `internal/service/reportes.go` (`Reportes.Sprint`, `ReporteSprint`, `FilaReporte`, `FilaDefecto`); la velocidad con `velocidadDe` de HU-35 (`panel.go`) | `internal/service/reportes_test.go` |
| Repositorios | sin cambios: alcanzan los listados de sprints, historias, foto, tareas, registros y defectos (T2) | — |
| Pantalla | `internal/server/reportes.go`, `web/templates/paginas/reporte_sprint.html`, botón en `sprint.html` y `d-print-none` en `base.html` | `internal/server/reportes_test.go` |
| Aplicación | `app/app.go` arma `Reportes` con los cierres, las horas y los defectos | `app/reportes_test.go` (memoria y PostgreSQL) |

| Criterio | Escenarios BDD | Tests |
|---|---|---|
| CA-37.1 Historias, estimaciones, esfuerzo, métricas y defectos | Las métricas del reporte de un sprint cerrado · Las historias del reporte con sus horas · El reporte conserva la foto del cierre · Reporte de un sprint que todavía no se cerró · Reporte de un sprint sin historias ni defectos | `TestReporteDeSprint_Cerrado`, `TestReporteDeSprint_ConservaLaFoto`, `TestReporteDeSprint_Abierto`, `TestReporteDeSprint_VelocidadHastaEseSprint`, `TestReporteDeSprint_SinHistorias`, `TestReporteDeSprint_Defectos`, `TestReporteDeSprint_DefectosResueltosDespuesOReabiertos`, `TestReporteDeSprint_SinDefectos`, `TestReporteDeSprint_MuestraElReporte`, `TestReporteDeSprint_SeImprime`, `TestPaginaDelSprint_EnlazaAlReporte`, `TestReporteDeSprintAbierto_AvisaQueEsParcial`, `TestReporteDeSprint_MuestraLosDefectos`, `TestNewHandlerConBase_SinURLSeVeElReporteDelSprint`, `TestNewHandlerConBase_ConURLElReporteSobreviveAUnReinicio` |
| CA-37.2 Reutiliza `internal/metrics` | Las métricas del reporte de un sprint cerrado | `TestReporteDeSprint_Cerrado` (los valores son los de cada métrica), `TestReporteDeSprint_Defectos` |
| CA-37.3 Sprint inexistente → 404 | Pedir el reporte de un sprint que no existe | `TestReporteDeSprint_Errores`, `TestReporteDeSprint_FallanLosDefectos`, `TestReporteDeSprint_NoExiste` |
