# SPEC-31 · Velocidad del equipo

**Historia:** HU-31 (#9) · **Responsable:** Bravo, Carolina · **Sprint:** 1 · **Estado:** Implementada

> Especificación escrita antes del código (SDD). Usa la plantilla de Mariano (`specs/_plantilla.md`) y la fórmula M3 de `docs/metricas.md`.

## 1. Objetivo

Que el equipo sepa cuántos story points termina, en promedio, por sprint, para decidir cuánto comprometer en el próximo sprint (lo usa el Agile Enabler en la Planning y se muestra en el dashboard).

## 2. Entradas

| Dato | Tipo | Obligatorio | Validación |
|---|---|---|---|
| Sprints del proyecto | lista de sprints (número, estado, SP completados) | Sí (puede estar vacía) | SP completados ≥ 0 |
| Estado de cada sprint | `Planificado`, `Activo` o `Cerrado` | Sí | Uno de los tres valores |
| Ventana N | número entero | No (por defecto 3) | N > 0 |

## 3. Salidas esperadas

- **Velocidad:** número con decimales, que en pantalla se muestra redondeado a 1 decimal (ej.: 16,5).
- **Mensaje:** "Aún no hay sprints cerrados" cuando no hay datos para calcular.

## 4. Reglas de negocio

- RN1: Solo cuentan los sprints en estado **Cerrado**; los activos y planificados se ignoran.
- RN2: Se toman los **últimos N** sprints cerrados, es decir, los de mayor número de sprint.
- RN3: Velocidad = suma de SP completados de esos sprints ÷ **k**, donde k = cantidad de sprints que se tomaron (k = N si hay N o más cerrados; si hay menos, k = los que haya).
- RN4: El orden en que llegan los sprints no cambia el resultado.
- RN5: Para un sprint cerrado se usan los SP guardados al cerrarlo ("foto del cierre", HU-14).

## 5. Restricciones

- Cálculo puro en `internal/metrics` (sin base de datos ni HTTP), así se prueba con TDD sin depender de nadie.
- No se divide nunca por cero.
- Internamente `float64`; el redondeo es solo para mostrar.

## 6. Casos límite

| Caso | Resultado |
|---|---|
| Sin sprints cerrados (lista vacía o solo activos/planificados) | 0 y el mensaje "Aún no hay sprints cerrados" |
| Menos sprints cerrados que N (ej.: 10 y 20 con N = 3) | promedio de los que hay: 15 |
| Exactamente N sprints cerrados | promedio de los N |
| Más de N (ej.: 10, 20, 30, 40 con N = 3) | promedio de los últimos 3: (20 + 30 + 40) / 3 = 30 |
| N = 1 (ej.: 12 y 8, el último es 8) | 8 |
| Sprint cerrado con 0 SP completados | cuenta como 0 en el promedio |

## 7. Condiciones de error

| Situación | Error | Mensaje al usuario |
|---|---|---|
| Ventana N = 0 o negativa | `ErrVentanaInvalida` | "la ventana de sprints debe ser mayor a cero" |
| Algún sprint con SP completados negativos | `ErrSPNegativos` | "los story points no pueden ser negativos" |

## 8. Criterios de aceptación

- CA-31.1: Calcula el promedio de SP completados de los últimos 3 sprints cerrados.
- CA-31.2: Si hay menos sprints cerrados que la ventana, promedia los que haya.
- CA-31.3: Sin sprints cerrados devuelve 0 y avisa "Aún no hay sprints cerrados".
- CA-31.4: No cuenta sprints activos ni planificados.
- CA-31.5: Una ventana de 0 o negativa devuelve error.

## Trazabilidad

Issue: #9 · Métrica: M3 en `docs/metricas.md` · Feature: `features/hu-31-velocidad.feature` · Tests: `internal/metrics/velocity_test.go` · Código: `internal/metrics/velocity.go` (`CalcularVelocidad`) · Pasos BDD: `features/hu31_steps_test.go`
