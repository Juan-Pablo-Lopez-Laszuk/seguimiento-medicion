# SPEC-34 · Defectos detectados y resueltos

**Historia:** HU-34 (#12) · **Responsable:** Bravo, Carolina · **Sprint:** 3 · **Estado:** Borrador

> Especificación escrita antes del código (SDD). Usa las métricas M8 y M9 de `docs/metricas.md`.

## 1. Objetivo

Saber cuántos defectos se detectaron y cuántos se resolvieron en cada sprint y en el proyecto, para medir la calidad.
Lo usan el panel (HU-35) y el reporte de sprint (HU-37).

## 2. Entradas

| Dato | De dónde sale |
|---|---|
| Proyecto | id (viene en la URL); tiene que existir |
| Defectos del proyecto | HU-27, con su estado y sus sprints de detección y de resolución (HU-28) |
| Severidad (filtro) | `?severidad=` en la URL: `Crítica`, `Alta`, `Media` o `Baja`. Vacío o cualquier otro valor = todas |

## 3. Salidas esperadas

- En la **página de defectos** (`/proyectos/{id}/defectos`), arriba de la lista, un **resumen**:
  - Un selector de severidad ("Todas" por defecto) que filtra el resumen.
  - Una fila por sprint con los detectados y los resueltos en ese sprint.
  - El total del proyecto: detectados, resueltos y abiertos.

## 4. Reglas de negocio

- RN1: Detectados (sprint S) = defectos cuyo sprint de detección es S (M8, CA-34.1).
- RN2: Resueltos (sprint S) = defectos en estado Resuelto o Cerrado cuyo sprint de resolución es S (M9, CA-34.1).
  Un defecto reabierto no cuenta como resuelto.
- RN3: Total del proyecto: todos los detectados, todos los resueltos y los abiertos (detectados − resueltos) (CA-34.2).
- RN4: El filtro por severidad se aplica a las filas y al total (CA-34.2).

## 5. Restricciones

- Conteo puro en `internal/metrics/defects.go` (`ContarDefectos`), como el resto de las métricas. El caso de uso
  `Defectos.Resumen` arma las filas. Solo lee: no cambia nada. Sin migración nueva.

## 6. Casos límite

| Caso | Resultado |
|---|---|
| Sin defectos | Todo en 0 |
| Defecto Reabierto | Cuenta como detectado, no como resuelto |
| Severidad desconocida en la URL | Se muestran todas |

## 7. Condiciones de error

| Condición | Error | Mensaje / respuesta |
|---|---|---|
| Proyecto inexistente | `project.ErrNoEncontrado` | Página 404 |

## 8. Criterios de aceptación

- CA-34.1: Detectados por sprint de detección y resueltos por sprint de resolución.
- CA-34.2: Total del proyecto y filtro por severidad.

## Trazabilidad

Issue: #12 · Métricas: M8 y M9 · Feature: `features/hu-34-defectos.feature` · Pasos BDD: `features/hu34_steps_test.go`
