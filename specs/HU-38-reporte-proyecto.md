# SPEC-38 · Reporte de proyecto

**Historia:** HU-38 (#49) · **Responsable:** López, Juan Pablo · **Sprint:** 3 (adelantada del Sprint 4) · **Estado:** Borrador

> Especificación escrita antes del código (SDD). Junta en una página el resumen del proyecto, que ya calcula el panel
> (HU-35), y el reporte de cada sprint, que ya arma HU-37. Agrega la evolución de la velocidad (CA-38.2), que sale de
> la velocidad que cada reporte de sprint ya calcula con los sprints cerrados hasta ese (HU-37 RN4). No calcula nada
> nuevo. Es lo que HU-39 va a pasar a PDF.

## 1. Objetivo

Que el Agile Enabler tenga, en una sola página, todo lo que se le rinde al cliente sobre el proyecto: cómo está, cuánto
se avanzó, cómo evolucionó la velocidad del equipo y qué pasó en cada sprint. La página se puede imprimir (cada sprint
en una hoja nueva) y se lee bien en el celular.

## 2. Entradas

| Dato | Tipo | Obligatorio | Validación |
|---|---|---|---|
| Proyecto | id del proyecto (viene en la URL) | Sí | Tiene que existir |

No hay datos que cargue el usuario: es una consulta. La página es `/proyectos/{id}/reporte` y se llega con el botón
**Reporte del proyecto** de la página del proyecto (HU-04), al lado del botón **Panel**.

## 3. Salidas esperadas

Con el ejemplo de la sección 6, el reporte muestra, en este orden:

1. **Encabezado:** "Reporte del proyecto · Software Metrics", la descripción, las fechas (05/10/2026 al 30/11/2026) y
   el estado calculado (HU-04).
2. **Resumen del proyecto** (CA-38.1), en tarjetas, con los mismos valores que el panel (HU-35):

   | Métrica | Valor | De dónde sale |
   |---|---|---|
   | Sprints | 2 cerrados de 3 | los sprints del proyecto |
   | SP completados | 8 de 13 SP (61,5 %) | M1 y M2 del proyecto (HU-30), como la página del proyecto |
   | Velocidad del equipo | 4,0 SP por sprint (2 sprints cerrados) | M3 (HU-31) |
   | Horas estimadas y reales | 12 h estimadas · 11 h reales | M4 y M5 del proyecto (HU-33) |
   | Desviación | -1 h (-8,3 %) | M6 del proyecto (HU-33) |
   | Defectos | 3 detectados · 2 resueltos · 1 abierto | M8 y M9 del proyecto (HU-34) |

3. **Evolución de la velocidad** (CA-38.2): una fila por sprint cerrado, ordenada por número.

   | Sprint | SP planificados | SP completados | Cumplimiento | Velocidad hasta ese sprint |
   |---|---|---|---|---|
   | Sprint 1 | 13 | 5 | 38,5 % | 5,0 SP por sprint |
   | Sprint 2 | 8 | 3 | 37,5 % | 4,0 SP por sprint |

   Sin sprints cerrados: "Aún no hay sprints cerrados".

4. **Un bloque por sprint** (CA-38.1), ordenados por número, con **el mismo contenido que el reporte de ese sprint**
   (HU-37): encabezado, tarjetas, historias y defectos. Cada bloque tiene un enlace a su reporte de sprint. El sprint
   que no está cerrado lleva el aviso de reporte parcial de HU-37.
5. Un botón **Imprimir**: la impresión oculta la navegación y los botones, y arranca cada sprint en una hoja nueva.

## 4. Reglas de negocio

- RN1: **El bloque de cada sprint es su reporte de HU-37**, sin cambios: foto del cierre en los cerrados (HU-37 RN1),
  historias de hoy y aviso en los abiertos (HU-37 RN2). Así el reporte del proyecto y el de cada sprint no pueden dar
  distinto.
- RN2: **El resumen es el del panel** (HU-35): mismos valores, misma regla de "sin datos". Los totales del proyecto
  **no se suman de los bloques**: una historia que pasó por dos sprints aparece en los dos bloques con sus horas (HU-37
  RN5), y sumarlos la contaría dos veces. Por eso las horas y los defectos del proyecto salen de HU-33 y HU-34.
- RN3: **Evolución de la velocidad:** una fila por sprint cerrado, con los SP de su cierre (HU-14), su cumplimiento
  (HU-15) y la velocidad hasta ese sprint (HU-37 RN4: últimos 3 cerrados con número menor o igual). Los sprints
  abiertos no entran, porque la velocidad solo cambia al cerrar un sprint.
- RN4: **Sprints** = cantidad de sprints cerrados y cantidad total de sprints del proyecto.
- RN5: Un proyecto **sin sprints** muestra el resumen, "Aún no hay sprints cerrados" en la evolución y "Todavía no hay
  sprints en este proyecto." en lugar de los bloques.
- RN6: Es solo lectura: consultar el reporte no cambia nada.

## 5. Restricciones

- **Sin cálculos nuevos:** el resumen sale de `Panel.Ver` (HU-35), los bloques de `Reportes.Sprint` (HU-37) y la
  evolución de los campos `SP`, `CumplimientoTexto` y `Velocidad` que cada `ReporteSprint` ya trae.
- **Sin migración ni repositorios nuevos.**
- **Caso de uso:** `Reportes.Proyecto(ctx, proyectoID) (ReporteProyecto, error)`. Para no leer la base una vez por
  sprint, `Reportes.Sprint` se parte en dos: una parte que lee los listados del proyecto (sprints, historias,
  defectos) y otra que arma el reporte de un sprint con esos listados. `Proyecto` los lee una sola vez y arma todos
  los sprints. El reporte de un sprint sigue dando lo mismo (sus tests no cambian).
- **Pantalla:** el bloque de un sprint se saca de `reporte_sprint.html` a una plantilla compartida
  (`{{define "reporte-sprint"}}`), que usan las dos páginas. `@media print` con `break-before: page` en cada bloque.
- **Depende de HU-37 (#416):** la rama sale de la de HU-37. Cuando se mergee #416, el PR queda solo con lo de HU-38.
- **Sin gráfico** [A CONFIRMAR]: la evolución va en tabla, que se imprime y pasa a PDF (HU-39) sin cambios. Si HU-36
  (Carolina) trae un gráfico de velocidad, se puede sumar después.

## 6. Casos límite

Proyecto "Software Metrics" (05/10 al 30/11) con tres sprints:

- **Sprint 1** (12/10 al 18/10), cerrado con HU-1 Login (5 SP, completada), HU-2 Registro (3 SP, no) y HU-3 Perfil
  (5 SP, no): 5 de 13 SP.
- **Sprint 2** (19/10 al 25/10), cerrado con HU-2 Registro (3 SP, completada) y HU-3 Perfil (5 SP, no): 3 de 8 SP.
- **Sprint 3** (26/10 al 01/11), abierto, con HU-3 Perfil (5 SP).

Horas: HU-1 con 6 h estimadas y 8 h reales; HU-2 con 4 h y 3 h; HU-3 con 2 h estimadas y 0 h reales. Defectos: #1
Crítica de HU-1, detectado y resuelto en el Sprint 1; #2 Media de HU-2, detectado en el Sprint 1 y resuelto en el 2;
#3 Baja, detectado en el Sprint 2 y abierto.

| Caso | Resultado |
|---|---|
| Reporte del proyecto | Lo de la sección 3, con un bloque por cada uno de los 3 sprints |
| Bloque del Sprint 1 | Igual al reporte del Sprint 1: 5 de 13 SP, velocidad 5,0 (RN1) |
| Bloque del Sprint 3 | Aviso de reporte parcial; no aparece en la evolución (RN3) |
| HU-2 aparece en los bloques de los Sprints 1 y 2 | Las horas del resumen son las del proyecto: 12 h y 11 h, no 20 h y 14 h (RN2) |
| Proyecto sin sprints | Resumen con "sin datos", "Aún no hay sprints cerrados" y "Todavía no hay sprints en este proyecto." (RN5) |
| Proyecto con sprints pero ninguno cerrado | Evolución: "Aún no hay sprints cerrados"; los bloques con el aviso de parcial |
| Re-estimar HU-3 después de cerrar el Sprint 2 | El bloque y la fila del Sprint 2 siguen con 3 de 8 SP (foto del cierre) |

## 7. Condiciones de error

| Situación | Error | Mensaje al usuario |
|---|---|---|
| El proyecto no existe | 404 | "No encontramos ese proyecto." (el de HU-11) |
| Falla al leer la base | 500 | — |

## 8. Criterios de aceptación

- CA-38.1 Consolida todos los sprints con el mismo contenido que el reporte de sprint.
- CA-38.2 Incluye la evolución de la velocidad.

## Trazabilidad

Issue: #49 · Feature: `features/hu-38-reporte-proyecto.feature` · Pasos BDD: `features/hu38_steps_test.go`

Las tablas por capa y por criterio se completan con el código.
