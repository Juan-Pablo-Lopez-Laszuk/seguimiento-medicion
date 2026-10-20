# SPEC-15 · Consultar sprints anteriores

**Historia:** HU-15 (#47) · **Responsable:** López, Juan Pablo · **Sprint:** 3 · **Estado:** Implementada

> Especificación escrita antes del código (SDD). Usa la lista de sprints de `specs/HU-11-crear-sprint.md`, la página
> del sprint de `specs/HU-12-asignar-historias.md` y la foto del cierre de `specs/HU-14-cerrar-sprint.md` (RN3). El
> porcentaje sigue la regla de HU-04 (RN4: nunca se divide por cero) y el formato de HU-32 (1 decimal y coma).
> Revisada por Juan Pablo el 19/10: sin pantallas nuevas, cumplimiento por SP de la foto y detalle desde la foto.

## 1. Objetivo

Que el equipo vea cómo le fue en cada sprint ya cerrado: qué se propuso, cuántos story points planificó, cuántos
completó y qué porcentaje cumplió, y pueda entrar al detalle de cada uno para ver qué historias tenía y cuáles terminó.
Sirve para la Retrospectiva y para comparar sprints.

## 2. Entradas

| Dato | Tipo | Obligatorio | Validación |
|---|---|---|---|
| Proyecto | id del proyecto (viene en la URL) | Sí | Tiene que existir |
| Sprint (solo el detalle) | número del sprint (viene en la URL) | Sí | Tiene que existir en el proyecto |

No hay datos que cargue el usuario: es una consulta.

## 3. Salidas esperadas

- **Lista** (CA-15.1): la página de sprints del proyecto (`/proyectos/{id}/sprints`, de HU-11) suma dos columnas,
  **SP completados** y **Cumplimiento**. Queda así, ordenada por número:

  | Sprint | Sprint Goal | Inicio | Fin | Estado | SP completados | Cumplimiento |
  |---|---|---|---|---|---|---|
  | Sprint 1 | Gestionar el backlog | 12/10/2026 | 18/10/2026 | Cerrado | 5 de 13 | 38,5 % |
  | Sprint 2 | Planificar sprints | 19/10/2026 | 25/10/2026 | Cerrado | 6 de 8 | 75,0 % |
  | Sprint 3 | Métricas | 26/10/2026 | 01/11/2026 | Planificado | — | — |

  Un sprint que todavía no se cerró muestra "—" en las dos columnas (RN1).
- **Detalle** (CA-15.2): cada sprint de la lista ya tiene su enlace a la página del sprint
  (`/proyectos/{id}/sprints/{numero}`). Para un sprint **cerrado**, esa página muestra:
  - arriba, el resultado: "**5 de 13 SP completados · 38,5 % de cumplimiento · 1 de 3 historias**";
  - la tabla **"Historias al cerrar"** con la foto del cierre, ordenada por número: identificador (con enlace a la
    historia), título, los SP que tenía al cerrar y el resultado, **Completada** o **No completada**;
  - el aviso de que el sprint está cerrado y no se le pueden asignar historias, como ahora (HU-12).

  Para un sprint abierto, la página no cambia.

## 4. Reglas de negocio

- RN1: **Los sprints anteriores son los cerrados.** Solo ellos muestran SP completados y cumplimiento, porque su
  resultado ya no cambia. Los demás (el activo y los futuros) muestran "—": su resultado todavía se está armando, y
  el avance del activo ya se ve en su página.
- RN2: **Cumplimiento** = SP completados ÷ SP planificados × 100, con 1 decimal y coma decimal ("38,5 %"). Se calcula
  con los SP **de la foto del cierre** (`sprints.sp_planificados` y `sp_completados`, HU-14 RN3), no con las historias
  de hoy.
- RN3: Si el sprint cerrado tiene **0 SP planificados** (se cerró sin historias, o solo con historias de 0 SP), el
  cumplimiento es "—": nunca se divide por cero (como HU-04 RN4).
- RN4: La lista va **ordenada por número** de sprint, de menor a mayor.
- RN5: **El detalle muestra la foto del cierre**: todas las historias que tenía el sprint al cerrarse, también las que
  volvieron al backlog, con los SP que tenían en ese momento. El "1 de 3 historias" cuenta las filas de la foto
  marcadas como completadas sobre el total de la foto (la misma cuenta de HU-32).
- RN6: **El detalle no cambia** aunque después una historia de la foto se re-estime, entre a otro sprint o se complete
  en otro sprint: los SP y el resultado salen de la foto. El título sí es el de hoy (la foto no guarda títulos), así un
  título corregido se ve corregido.
- RN7: Es solo lectura: consultar no cambia nada.

## 5. Restricciones

- **Cálculo puro en `internal/metrics`:** el porcentaje de cumplimiento (RN2 y RN3) es una función sin base de datos
  ni HTTP, con su texto para pantalla, como `Porcentaje.Texto` de HU-32.
- **Sin migración ni cambios en los repositorios:** la lista usa `SPPlanificados` y `SPCompletados` de
  `sprint.Sprint`, que ya se leen (HU-14), y el detalle usa `Cierres.Foto` de HU-14. Los títulos salen de las
  historias del proyecto.
- **Casos de uso:** la lista es `Sprints.Resumir` (junto a `Listar` de HU-11) y el detalle es `Cierres.Detalle`
  (el caso de uso de HU-14, que ya tiene la foto). Así `Asignaciones` no cambia: la página del sprint pide el detalle
  solo si el sprint está cerrado.
- Un sprint cerrado que no tenga `sp_planificados` (por ejemplo, cargado a mano antes de HU-14) se muestra como sin
  resultado ("—"), igual que uno abierto.
- **Sin pantallas nuevas:** la lista es la de HU-11 y el detalle es la página del sprint de HU-12. Así no hay dos
  listas de sprints que mantener.
- Los repositorios devuelven los sprints en el orden en que se crearon, que es el orden de los números (HU-11). Igual
  el caso de uso ordena por número, para no depender de eso.

## 6. Casos límite

Proyecto con el Sprint 1 (12/10 al 18/10) cerrado con HU-1 (5 SP, completada), HU-2 (3 SP, no) y HU-3 (5 SP, no); el
Sprint 2 (19/10 al 25/10) cerrado con HU-4 (3 SP, completada), HU-5 (3 SP, completada) y HU-6 (2 SP, no); y el
Sprint 3 (26/10 al 01/11) sin cerrar.

| Caso | Resultado |
|---|---|
| La lista | Sprint 1: 5 de 13, 38,5 % · Sprint 2: 6 de 8, 75,0 % · Sprint 3: — y — |
| Detalle del Sprint 1 | 5 de 13 SP · 38,5 % · 1 de 3 historias; HU-1 Completada, HU-2 y HU-3 No completadas |
| HU-2 se re-estima en 8 y entra al Sprint 3 | El detalle del Sprint 1 sigue con HU-2 en 3 SP y No completada (RN6) |
| Sprint cerrado sin historias | Lista: 0 de 0 y —; detalle: "El sprint se cerró sin historias" (RN3) |
| Sprint cerrado con todo completado | 100,0 % |
| Sprint cerrado sin nada completado | 0 de 13, 0,0 % |
| Proyecto sin sprints cerrados | Todos los sprints con — en las dos columnas |
| Proyecto sin sprints | El mensaje de HU-11: "Todavía no hay sprints en este proyecto." |

## 7. Condiciones de error

| Situación | Error | Mensaje al usuario |
|---|---|---|
| El proyecto no existe | 404 | El de HU-11 |
| El sprint no existe en el proyecto | 404 | "No encontramos ese sprint." (el de HU-12) |
| Falla al leer la base | 500 | — |

## 8. Criterios de aceptación

- CA-15.1 Lista ordenada por número con goal, fechas, SP planificados/completados y % de cumplimiento.
- CA-15.2 Acceso al detalle de cada sprint.

## Trazabilidad

Issue: #47 · Feature: `features/hu-15-sprints-anteriores.feature` · Pasos BDD: `features/hu15_steps_test.go`

| Capa | Código | Tests |
|---|---|---|
| Métrica | `internal/metrics/cumplimiento.go` (`StoryPoints.Cumplimiento`, `CumplimientoTexto`) | `metrics/cumplimiento_test.go` |
| Caso de uso | `internal/service/anteriores.go` (`Sprints.Resumir`, `ResumenSprint`, `Cierres.Detalle`, `DetalleCierre`, `FilaFoto`) | `internal/service/anteriores_test.go` |
| Pantalla | `internal/server/sprints.go` (lista), `internal/server/planificacion.go` (detalle), `web/templates/paginas/sprints.html` y `sprint.html` | `internal/server/anteriores_test.go` |
| Aplicación | sin cambios: la página del sprint ya recibe `Cierres` (HU-14) | `app/anteriores_test.go` (memoria y PostgreSQL) |

| Criterio | Escenarios BDD | Tests |
|---|---|---|
| CA-15.1 Lista ordenada con goal, fechas, SP y cumplimiento | Ver los sprints anteriores con su resultado · Sprint cerrado sin historias · Sprint cerrado con todas sus historias completadas | `TestCumplimiento_CompletadosSobrePlanificados`, `TestCumplimiento_SinSPPlanificadosNoDividePorCero`, `TestResumirSprints_OrdenadosConElResultadoDeLosCerrados`, `TestResumirSprints_CerradoSinHistorias`, `TestResumirSprints_Errores`, `TestListaDeSprints_MuestraElResultadoDeLosCerrados`, `TestSprintCerradoSinHistorias` |
| CA-15.2 Detalle de cada sprint | Ver el detalle de un sprint cerrado · El detalle conserva la foto aunque la historia cambie después · Consultar un sprint que no existe | `TestDetalleDeCierre_MuestraLaFoto`, `TestDetalleDeCierre_NoCambiaDespuesDelCierre`, `TestDetalleDeCierre_SprintAbierto`, `TestDetalleDeCierre_Errores`, `TestPaginaDeSprintCerrado_MuestraLaFoto`, `TestPaginaDeSprintAbierto_NoTieneLaFoto`, `TestNewHandlerConBase_SinURLSeConsultaElSprintAnterior`, `TestNewHandlerConBase_ConURLElSprintAnteriorSobreviveAUnReinicio` |
