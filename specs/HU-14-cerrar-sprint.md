# SPEC-14 · Cerrar sprint

**Historia:** HU-14 (#46) · **Responsable:** López, Juan Pablo · **Sprint:** 3 · **Estado:** Implementada

> Especificación escrita antes del código (SDD). Usa el sprint activo de `specs/HU-04-estado-proyecto.md` (RN2), la
> página del sprint de `specs/HU-12-asignar-historias.md`, la regla de estados de `specs/HU-08-estado-historia.md`
> (En sprint → Pendiente y En progreso → Pendiente "al cerrar el sprint"), el cálculo de
> `specs/HU-30-sp-planificados-completados.md` y la decisión del #80 (la foto del cierre). Revisada por Juan
> Pablo el 19/10: se puede cerrar antes de la fecha de fin (RN2), cualquiera puede cerrar hasta TEC-05 y una historia
> que volvió al backlog ya no se elimina (sección 5).

## 1. Objetivo

Que el Agile Enabler cierre el sprint activo y quede guardado su resultado: cuántos story points se planificaron y
cuántos se completaron, y qué historias participaron. Con esa foto se calculan la velocidad (HU-31), los sprints
anteriores (HU-15), el reporte del sprint (HU-37) y el panel (HU-35), aunque después las historias cambien.

## 2. Entradas

| Dato | Tipo | Obligatorio | Validación |
|---|---|---|---|
| Proyecto | id del proyecto (viene en la URL) | Sí | Tiene que existir |
| Sprint | número del sprint (viene en la URL) | Sí | Tiene que existir y ser el sprint activo (RN1) |
| Hoy | fecha | — | No la carga el usuario: es la fecha de Argentina (HU-04 RN6), para saber cuál es el sprint activo |

## 3. Salidas esperadas

- En la página del sprint activo (HU-12) hay un botón **Cerrar sprint**, que lleva a una **pantalla de confirmación**
  (`/proyectos/{id}/sprints/{numero}/cerrar`) con lo que va a pasar:
  - "Se completaron **8 de 13 SP** (1 de 3 historias)".
  - La lista de historias que **vuelven al backlog como Pendientes**, con el estado en que están (por ejemplo
    "HU-2 · Registro (En progreso)").
  - El aviso "Un sprint cerrado no se puede reabrir ni modificar".
  - Si se cierra antes de la fecha de fin: "El sprint termina el 25/10/2026: lo estás cerrando 3 días antes".
- Mostrar la confirmación no cambia nada. Recién el botón **Cerrar sprint** de esa pantalla (un `POST`) cierra el
  sprint y vuelve a su página con el aviso "Sprint 1 cerrado: 8 de 13 SP completados. 2 historias volvieron al
  backlog".
- La página de un sprint cerrado muestra el estado **Cerrado** y los SP de la foto ("8 de 13 SP completados"), sin
  botones para asignar, quitar, empezar ni completar. La lista de sprints (HU-11) muestra el estado Cerrado.
- El detalle de lo que tenía planificado un sprint cerrado, incluidas las historias que volvieron al backlog, lo
  muestra HU-15 con la foto del cierre.

## 4. Reglas de negocio

- RN1: **Solo se cierra el sprint activo** (CA-14.1): el no cerrado de menor número cuya fecha de inicio ya llegó
  (HU-04 RN2). Un sprint vencido sin cerrar sigue activo, así que se puede cerrar tarde. Un sprint futuro, o uno
  posterior a otro que sigue abierto, no se cierra.
- RN2: **Se puede cerrar antes de la fecha de fin.** La confirmación avisa cuántos días faltaban.
- RN3: **La foto del cierre** (CA-14.3, decisión del #80). Al cerrar se guarda:
  - en `sprints`: el estado **Cerrado**, `sp_planificados` y `sp_completados`;
  - en `sprint_historias`: una fila por cada historia que estaba en el sprint, con sus **story points en ese momento**
    y si estaba **Hecha** (`completada`).

  Los SP se calculan con `metrics.CalcularStoryPoints` de HU-30: planificados = Σ SP de las historias del sprint;
  completados = Σ SP de las Hechas. Una historia que después se re-estime o entre a otro sprint no cambia la foto.
- RN4: **Las historias no completadas vuelven al backlog** (CA-14.2): las que están En sprint o En progreso pasan a
  **Pendiente** y quedan **sin sprint**, con la regla de HU-08 (las dos transiciones a Pendiente "al cerrar el sprint").
  Conservan sus criterios (cumplidos o no), sus tareas y su estimación. Las Hechas se quedan Hechas y en el sprint.
- RN5: **Un sprint cerrado no se reabre ni se modifica** (CA-14.4). No hay forma de reabrirlo, y las acciones sobre un
  sprint cerrado ya se rechazan: asignar (HU-12, "el Sprint N está cerrado: no se le pueden asignar historias"),
  quitar (solo se quitan historias En sprint, y en un sprint cerrado solo quedan las Hechas), y empezar y completar
  (HU-13, solo el sprint activo). Cerrarlo otra vez es un error.
- RN6: **Todo o nada.** El estado del sprint, sus SP, la foto y las historias que vuelven al backlog se guardan juntos:
  si algo falla, no se guarda nada.
- RN7: Cerrar un sprint **sin historias** vale: queda con 0 de 0 SP y la foto vacía.
- RN8: Después del cierre, el siguiente sprint sin cerrar pasa a ser el activo **cuando llega su fecha de inicio**
  (HU-04 RN2). Si se cierra antes de tiempo, puede haber unos días sin sprint activo.

## 5. Restricciones

- **Dominio** (`internal/domain/sprint`): una función pura `Cerrar` que recibe el sprint, los sprints del proyecto, las
  historias del sprint y hoy, y devuelve el sprint cerrado, la foto y las historias que vuelven al backlog (o el error de
  la regla). Las transiciones a Pendiente son las de HU-08 (`Historia.CambiarEstado`), no se repiten.
- **Repositorio:** `Cierres`, con `Cerrar` (guarda todo junto, RN6) y `Foto` (lee la foto de un sprint). En PostgreSQL,
  en **una transacción** (`UPDATE sprints`, `INSERT INTO sprint_historias`, `UPDATE historias`); el `UPDATE` de
  `sprints` solo cambia un sprint que no está cerrado, así dos cierres al mismo tiempo no pisan la foto. En memoria,
  `memory.Cierres` valida todo antes de cambiar los repositorios de sprints e historias.
- **Sin migración:** `sprints.sp_planificados`, `sprints.sp_completados` y `sprint_historias` ya están en la 0001.
  `sprint.Sprint` suma `SPPlanificados` (ya tiene `SPCompletados`, que lee la velocidad).
- **HU-06:** `EnSprintCerrado` del repositorio en memoria hoy devuelve siempre `false` porque no se podía cerrar un
  sprint. HU-14 lo completa para que haga lo mismo que el de PostgreSQL: una historia es de un sprint cerrado si su
  sprint está cerrado **o** si aparece en la foto de un sprint cerrado (decisión 2 de la Planning del Sprint 3). Así,
  una historia que volvió al backlog no se puede eliminar, porque la foto la referencia.
- **Sin autenticación:** "el Agile Enabler" es un rol de la historia; hasta TEC-05 (Sprint 4) cualquiera que use la
  app puede cerrar.
- La velocidad (HU-31) y el aviso de HU-12 ya leen `SPCompletados` de los sprints cerrados: empiezan a dar valores sin
  cambiar su código.
- La página del sprint (HU-12) muestra, para un sprint cerrado, los SP del cierre en lugar de sumar las historias que
  le quedan (las no completadas ya volvieron al backlog). `buscarSprint` del caso de uso se comparte con el cierre.

## 6. Casos límite

Proyecto con el Sprint 1 activo, del 12/10 al 18/10, y el Sprint 2 del 19/10 al 25/10. El Sprint 1 tiene HU-1 (5 SP,
Hecha), HU-2 (3 SP, En progreso) y HU-3 (5 SP, En sprint).

| Caso | Resultado |
|---|---|
| Cerrar el Sprint 1 el 18/10 | Cerrado, 13 SP planificados y 5 completados; HU-2 y HU-3 vuelven a Pendiente y sin sprint; HU-1 sigue Hecha en el Sprint 1 |
| La foto del Sprint 1 | 3 filas: HU-1 (5 SP, completada), HU-2 (3 SP, no), HU-3 (5 SP, no) |
| Después del cierre, HU-2 se re-estima en 8 | La foto sigue diciendo 3 SP y el Sprint 1 sigue con 13 planificados |
| Cerrar el Sprint 1 el 20/10 (vencido sin cerrar) | Se cierra igual (RN1) |
| Cerrar el Sprint 1 el 16/10 | Se cierra; la confirmación avisa "lo estás cerrando 2 días antes" (RN2) |
| Sprint sin historias | Cerrado con 0 de 0 SP (RN7) |
| Todas las historias Hechas | Planificados = completados; ninguna vuelve al backlog |
| Ninguna historia Hecha | 0 SP completados; todas vuelven al backlog |
| Historia estimada en 0 SP | Entra en la foto con 0 SP |
| Después de cerrar el Sprint 1 | La velocidad es 5 SP (HU-31, con un solo sprint cerrado) |
| Cerrar el Sprint 1 el 16/10, con el Sprint 2 desde el 19/10 | Del 16/10 al 18/10 no hay sprint activo; el 19/10 lo es el Sprint 2 (RN8) |

## 7. Condiciones de error

| Situación | Error | Mensaje al usuario |
|---|---|---|
| El sprint ya está cerrado | `ErrYaCerrado` (409) | "el Sprint 1 ya está cerrado" |
| El sprint no es el activo (futuro, o hay uno anterior abierto) | `ErrNoEsElActivo` (409) | "solo se cierra el sprint activo (el Sprint 1)", o "solo se cierra el sprint activo: hoy no hay ninguno" |
| El proyecto o el sprint no existen | 404 | "No encontramos ese sprint." |
| Falla al guardar | 500 | No se guarda nada (RN6) |

Los errores se muestran en la pantalla de confirmación (409), que en ese caso no ofrece el botón **Cerrar sprint**.

## 8. Criterios de aceptación

- CA-14.1 Sólo se cierra el sprint activo.
- CA-14.2 Las historias no completadas vuelven al backlog como Pendientes.
- CA-14.3 Se guardan los SP planificados y completados (foto del cierre).
- CA-14.4 Un sprint cerrado no se reabre ni se modifica.

## Trazabilidad

Issue: #46 · Feature: `features/hu-14-cerrar-sprint.feature` · Pasos BDD: `features/hu14_steps_test.go`

| Capa | Código | Tests |
|---|---|---|
| Dominio | `internal/domain/sprint/cerrar.go` (`Cerrar`, `Cierre`, `FotoHistoria`, `ErrYaCerrado`, `ErrNoEsElActivo`) y `Sprint.SPPlanificados` | `sprint/cerrar_test.go` |
| Repositorios | `internal/store/{memory,postgres}/cierres.go` (`Cerrar`, `Foto`); `sp_planificados` en `postgres/sprints.go`; `EnSprintCerrado` en `memory/historias.go` | `memory/cierres_test.go`, `postgres/cierres_test.go` |
| Caso de uso | `internal/service/cierres.go` (`Cierres.Preparar`, `Cerrar`); en `asignaciones.go`, `Ver` de un sprint cerrado | `internal/service/cierres_test.go` |
| Pantalla | `internal/server/cierres.go`, `web/templates/paginas/sprint_cerrar.html` y el botón en `sprint.html` | `internal/server/cierres_test.go` |
| Aplicación | `app/app.go` (memoria y PostgreSQL) | `app/cierres_test.go` |

| Criterio | Escenarios BDD | Tests |
|---|---|---|
| CA-14.1 Solo el sprint activo | Cerrar un sprint vencido · Cerrar antes de la fecha de fin · No se cierra un sprint que no es el activo · Hoy no hay sprint activo | `TestCerrar_SprintActivoEnCualquierDia`, `TestCerrar_Errores`, `TestCerrar_IgnoraLaHora`, `TestPrepararCierre_UltimoDiaYVencido`, `TestCerrarSprint_Errores` (servicio y pantalla), `TestCerrarSprint_BotonSoloEnElSprintActivo`, `TestConfirmacionDeCierre_MuestraLoQueVaAPasar` |
| CA-14.2 Las no completadas vuelven al backlog | Las historias no completadas vuelven al backlog | `TestCerrar_LasNoCompletadasVuelvenAlBacklog`, `TestCierres_Cerrar` (memoria y PostgreSQL), `TestCerrarSprint_GuardaElCierre`, `TestCerrarSprint_LaNoCompletadaVuelveAlBacklog`, `TestNewHandlerConBase_SinURLSeCierraElSprint` |
| CA-14.3 Foto con los SP planificados y completados | Cerrar el sprint activo guarda su resultado · La foto no cambia después del cierre · La velocidad usa el sprint cerrado · Sprint sin historias | `TestCerrar_GuardaLaFotoYLosSP`, `TestCerrar_LaFotoNoComparteLosStoryPoints`, `TestCerrar_CasosLimite`, `TestCierres_Cerrar`, `TestVerSprint_CerradoMuestraLosSPDelCierre`, `TestCerrarSprint_VuelveALaPaginaDelSprintCerrado`, `TestNewHandlerConBase_ConURLElCierreSobreviveAUnReinicio` |
| CA-14.4 No se reabre ni se modifica | Un sprint cerrado no se cierra otra vez · A un sprint cerrado no se le asignan historias | `TestCerrar_Errores`, `TestCierres_Rechazos_NoGuardanNada` (memoria y PostgreSQL), `TestCerrarSprint_Errores` |
| RN6 Todo o nada · HU-06 RN6 | — | `TestCierres_Rechazos_NoGuardanNada`, `TestHistorias_EnSprintCerrado`, `TestCerrar_DatosInconsistentes` |
