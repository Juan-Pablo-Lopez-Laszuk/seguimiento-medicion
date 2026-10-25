# SPEC-36 · Gráficos de métricas

**Historia:** HU-36 (#14) · **Responsable:** Bravo, Carolina · **Sprint:** 4 · **Estado:** Borrador

> Especificación escrita antes del código (SDD). Usa las métricas que ya existen (M1 a M9 de `docs/metricas.md`), el
> reporte de cada sprint (HU-37, con la foto del cierre de HU-14) y la fecha en que se completó cada historia (HU-13).
> Agrega el burndown de la sección 5 de `docs/metricas.md`. Tareas: #207 a #213.

## 1. Objetivo

Ver las métricas del proyecto en gráficos, para ver las tendencias de un vistazo.

## 2. Entradas

| Dato | De dónde sale | Validación |
|---|---|---|
| Proyecto | id (viene en la URL) | Tiene que existir |
| Sprint del burndown | `?sprint=` (número) | Sprint del proyecto; otro valor o vacío = el sprint activo y, si no hay, el último |
| Métricas por sprint | El reporte de cada sprint de HU-37 (`Reportes.Proyecto`, HU-38) | — |
| Defectos | Los defectos del proyecto (HU-27, HU-28) | — |
| Fecha en que se completó cada historia | `CompletadaEn` de la historia (HU-13) | — |

## 3. Salidas esperadas

**Página de gráficos** (`/proyectos/{id}/graficos`), con botón en la vista del proyecto. Cinco gráficos (CA-36.1):

| Gráfico | Qué muestra |
|---|---|
| Velocidad por sprint | Por cada sprint cerrado, los SP completados (M2) y la velocidad hasta ese sprint (M3) |
| SP planificados vs completados | Por cada sprint, M1 y M2: los de la foto si está cerrado y los de hoy si no |
| Horas estimadas vs reales | Por cada sprint, M4 y M5 de sus historias |
| Defectos por severidad | Por severidad (Crítica, Alta, Media y Baja), los detectados (M8) y los que siguen abiertos |
| Burndown | Del sprint elegido: por cada día, los SP que faltan y la línea ideal |

Debajo de cada gráfico, una tabla "Ver los datos" con los mismos números, que se ve sin JavaScript y sirve para
imprimir y para lectores de pantalla.

**Endpoint JSON** (`/proyectos/{id}/graficos/datos`, con el mismo `?sprint=`): los datos de los cinco gráficos (CA-36.2).
La página los pide con JavaScript y dibuja los gráficos con Chart.js.

```json
{
  "velocidad":   {"sprints": ["Sprint 1"], "completados": [8], "velocidad": [8]},
  "storyPoints": {"sprints": ["Sprint 1", "Sprint 2"], "planificados": [13, 13], "completados": [8, 10]},
  "horas":       {"sprints": ["Sprint 1", "Sprint 2"], "estimadas": [10, 8], "reales": [12, 5]},
  "defectos":    {"severidades": ["Crítica", "Alta", "Media", "Baja"], "detectados": [1, 2, 0, 1], "abiertos": [1, 1, 0, 1]},
  "burndown":    {"sprint": "Sprint 2", "dias": ["12/10", "13/10"], "ideal": [13, 10.8], "pendientes": [13, 5]}
}
```

## 4. Reglas de negocio

- RN1: Los gráficos no calculan nada nuevo salvo el burndown: usan el reporte de cada sprint (HU-37), así un gráfico y
  el reporte no pueden dar distinto. Un sprint cerrado usa la foto de su cierre (RG-6).
- RN2: Burndown (`docs/metricas.md`, sección 5): para cada día del sprint, **SP pendientes = SP planificados − SP de
  las historias completadas hasta ese día** (por `CompletadaEn`, HU-13). Una historia completada antes de que empiece
  el sprint cuenta desde el primer día. Los pendientes nunca bajan de 0.
- RN3: La línea ideal baja en línea recta de los SP planificados (primer día) a 0 (último día), con 1 decimal.
- RN4: Los días que todavía no pasaron (después de hoy, en Argentina) no tienen SP pendientes: en el JSON van en
  `null` y en el gráfico no se dibujan.
- RN5: El burndown es del sprint del `?sprint=`. Si no viene o no existe, es el del sprint activo y, si no hay activo,
  el del último sprint. Sin sprints, `burndown` es `null` y la página dice "Sin sprints todavía".
- RN6: Solo lee: no cambia nada.

## 5. Restricciones

- El burndown es una función pura en `internal/metrics` (`CalcularBurndown`), con TDD (RG-1) y SP negativos como
  error (`ErrSPNegativos`, RG-4).
- Caso de uso `service.Graficos` (`Datos`), que usa `Reportes.Proyecto` (HU-38) y los defectos de HU-27.
- Decisión 4 de la Planning del Sprint 4: **Chart.js 4.4.1 por CDN**, sin paso de build, como Bootstrap. Si el CDN no
  carga, quedan las tablas con los datos.
- Los gráficos se ajustan al ancho de la pantalla (celular) y se imprimen.
- Sin migración nueva.

## 6. Casos límite

| Caso | Resultado |
|---|---|
| Proyecto sin sprints | Las series vacías y `burndown` en `null` |
| Sin sprints cerrados | La velocidad vacía (aviso "Todavía no hay sprints cerrados") |
| Sprint de un solo día | El ideal es 0 en ese día |
| Más SP completados que planificados | Los pendientes quedan en 0 |
| `?sprint=` que no existe | Se ignora (RN5) |

## 7. Condiciones de error

| Condición | Error | Respuesta |
|---|---|---|
| Proyecto inexistente | `project.ErrNoEncontrado` | 404 en la página y en el JSON (`{"error": "No encontramos ese proyecto."}`) |
| Sprint que termina antes de empezar | `metrics.ErrSprintSinDias` | No pasa con datos válidos (HU-11 lo impide); 500 |

## 8. Criterios de aceptación

- CA-36.1: Velocidad por sprint, SP planificados vs completados, horas estimadas vs reales, defectos por severidad y
  burndown.
- CA-36.2: Los datos salen de un endpoint JSON con tests.

## Trazabilidad

Issue: #14 · Feature: `features/hu-36-graficos.feature` · Pasos BDD: `features/hu36_steps_test.go`

(Se completa al implementar.)
