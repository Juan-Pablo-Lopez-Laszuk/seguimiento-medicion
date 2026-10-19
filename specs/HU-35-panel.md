# SPEC-35 · Panel de indicadores del proyecto

**Historia:** HU-35 (#13) · **Responsable:** Bravo, Carolina · **Sprint:** 3 · **Estado:** Borrador

> Especificación escrita antes del código (SDD). Junta métricas que ya existen: estado del proyecto (HU-04), velocidad
> (M3, HU-31 con los sprints cerrados de HU-14), story points completados (M2, HU-30), desviación de horas (M6, HU-33)
> y defectos (M8 y M9, HU-34).

## 1. Objetivo

Ver en una sola pantalla cómo viene el proyecto, para decidir rápido.

## 2. Entradas

| Dato | De dónde sale |
|---|---|
| Proyecto | id (viene en la URL); tiene que existir |
| Estado, sprint activo y story points | Caso de uso de HU-04 (`EstadoProyectos.Consultar`) |
| Velocidad | Los SP completados de los últimos 3 sprints cerrados (HU-31, HU-14) |
| Desviación de horas | Total del proyecto de HU-33 |
| Defectos abiertos | Total del proyecto de HU-34 (detectados − resueltos) |

## 3. Salidas esperadas

**Panel** (`/proyectos/{id}/panel`), con botón en la vista del proyecto. Seis tarjetas (CA-35.1):

| Tarjeta | Con datos | Sin datos (CA-35.2) |
|---|---|---|
| Estado del proyecto | Planificado, En curso o Finalizado | — (siempre hay) |
| Sprint activo | "Sprint N" y los días que le quedan | "sin datos" |
| Velocidad | "N SP por sprint" y de cuántos sprints cerrados sale | "sin datos" (no hay sprints cerrados) |
| % completado | % de SP completados sobre planificados | "sin datos" (no hay historias estimadas) |
| Desviación de horas | "+2 h" y "+20,0 %" | "sin datos" (no hay horas estimadas ni reales) |
| Defectos abiertos | Cantidad, y de cuántos detectados | "sin datos" (no hay defectos) |

## 4. Reglas de negocio

- RN1: Cada tarjeta usa la métrica que ya existe; el panel no calcula nada nuevo.
- RN2: Una tarjeta sin datos muestra "sin datos" en lugar de 0, y la página nunca da error por falta de datos
  (CA-35.2, decisión 5 de la Planning del Sprint 3).
- RN3: Si hay horas reales pero no estimadas, la desviación muestra las horas y el % dice "no aplicable" (HU-33).

## 5. Restricciones

- Caso de uso `service.Panel`, que usa los casos de uso de HU-04, HU-33 y HU-34. Solo lee: no cambia nada.
- Sin migración nueva.

## 6. Casos límite

| Caso | Resultado |
|---|---|
| Proyecto recién creado, sin nada | Estado Planificado y las otras cinco tarjetas en "sin datos" |
| Hay defectos pero todos resueltos | Defectos abiertos = 0 (sí hay datos) |
| Un solo sprint cerrado | La velocidad es la de ese sprint |

## 7. Condiciones de error

| Condición | Error | Mensaje / respuesta |
|---|---|---|
| Proyecto inexistente | `project.ErrNoEncontrado` | Página 404 |

## 8. Criterios de aceptación

- CA-35.1: Tarjetas con estado, sprint activo, velocidad, % completado, desviación de horas y defectos abiertos.
- CA-35.2: Si no hay datos, se ve vacío y sin errores.

## Trazabilidad

Issue: #13 · Feature: `features/hu-35-panel.feature` · Pasos BDD: `features/hu35_steps_test.go`
