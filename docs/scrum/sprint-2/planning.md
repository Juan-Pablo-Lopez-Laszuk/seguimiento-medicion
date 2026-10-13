# Sprint Planning · Sprint 2 · 12/10/2026

Participantes: López, Piastrellini, Bravo · Facilitador: Juan Pablo

> Armada el 12/10 a partir del [refinamiento](refinamiento.md), el tablero y la
> [Retro del Sprint 1](../sprint-1/retro.md). El compromiso ya está cargado en el tablero (Sprint Backlog). Lo marcado
> [A CONFIRMAR] lo cierra el equipo comentando en el PR de esta acta.

## Fechas

El Sprint 1 se estiró hasta el lunes 12/10, así que el Sprint 2 va del **martes 13/10 al domingo 18/10** (6 días, uno
menos que un sprint normal). El tablero lo tiene del 12/10 al 18/10, así que las fechas no cambian.

## Sprint Goal

> Trabajar el backlog y el sprint desde la app: ver el estado del proyecto, completar las historias (criterios,
> estimación y estado), asignarlas al sprint y marcarlas como completadas, todo con un diseño base común; y empezar a
> registrar tareas, esfuerzo y defectos.

[A CONFIRMAR]

## Capacidad y velocidad

- Velocidad del Sprint 1: **39 SP** en 8 días (con el día extra). Este sprint tiene 6 días: a ese ritmo entran unos
  **30 SP**.
- Hay **69 SP** con Sprint = 2 en el tablero, así que se recortan más de la mitad. Se proponen **33 SP**, algo por
  encima del ritmo, porque ninguna historia depende ya de TEC-03 ni de HU-05.
- Criterio del recorte (acción de la Retro: el camino crítico primero):
  1. Primero lo que destraba a otros: HU-07, HU-16 y HU-08 (Mariano) habilitan HU-12 y HU-13 (Juan Pablo), y TEC-07
     (Carolina) cambia la plantilla base que usan todas las pantallas.
  2. Después lo que no depende de nadie: HU-04 y HU-02 (Juan Pablo), y HU-25, HU-24 y HU-27 (Carolina).
  3. Lo que queda al final de la cadena (HU-14 y HU-15) y lo que es otra épica (Planning Poker y autenticación) pasa al
     Sprint 3.

## Historias del sprint

### Comprometidas

| Orden | Historia | SP | Responsable | Depende de | Por qué en este orden |
|---|---|---|---|---|---|
| 1 | TEC-07 · Diseño base de la interfaz (#16) | 3 | Carolina | — | Toca `base.html`: si entra temprano, las demás pantallas del sprint ya nacen con el diseño nuevo y no hay conflictos |
| 1 | HU-07 · Gestionar criterios de aceptación (#20) | 2 | Mariano | HU-05 (hecha) | Camino crítico: HU-12 exige historias con criterios |
| 1 | HU-16 · Estimar historia con story points (#24) | 2 | Mariano | HU-05 (hecha) | Camino crítico: HU-12 exige historias estimadas |
| 1 | HU-04 · Consultar estado del proyecto (#42) | 3 | Juan Pablo | HU-01, HU-03, HU-05, HU-11 (hechas) | Spec lista; define el "sprint activo" que usan HU-13 y HU-14 |
| 2 | HU-08 · Cambiar estado de una historia (#21) | 3 | Mariano | HU-07 | Camino crítico: HU-13 usa su regla para pasar a Hecho |
| 2 | HU-02 · Modificar proyecto (#40) | 2 | Juan Pablo | HU-04 | Spec lista |
| 2 | HU-25 · Descomponer historia en tareas (#3) | 3 | Carolina | HU-03, HU-05 (hechas) | Base de HU-24 |
| 3 | HU-12 · Asignar historias a un sprint (#44) | 5 | Juan Pablo | HU-07, HU-16, HU-08 | Sale cuando Mariano mergea HU-07 y HU-16 |
| 3 | HU-24 · Registrar esfuerzo realizado (#2) | 3 | Carolina | HU-03, HU-25 | — |
| 3 | HU-06 · Editar y eliminar ítem (#19) | 2 | Mariano | HU-05 (hecha) | No bloquea a nadie: va después del camino crítico |
| 4 | HU-13 · Registrar historias completadas (#45) | 2 | Juan Pablo | HU-12, HU-08, HU-04 | Última de la cadena del sprint |
| 4 | HU-27 · Registrar defecto (#5) | 3 | Carolina | HU-05 (hecha) | — |

**Total comprometido: 33 SP** · Juan Pablo 12 · Mariano 9 · Carolina 12.

Mariano tiene menos SP a propósito: sus tres primeras historias son el camino crítico y tienen que estar en `main`
para el **miércoles 14/10** si Juan Pablo va a llegar con HU-12 y HU-13.

### Si sobra tiempo (no comprometidas)

| Historia | SP | Responsable | Condición |
|---|---|---|---|
| HU-09 · Priorizar el backlog (#22) | 3 | Mariano | Si HU-07, HU-16, HU-08 y HU-06 están en `main` |
| HU-14 · Cerrar sprint (#46) | 5 | Juan Pablo | Si HU-12 y HU-13 están en `main` antes del viernes 16/10 |
| HU-28 · Gestionar el estado de un defecto (#6) | 3 | Carolina | Después de HU-27 |

Si se empieza alguna, se pasa a *Sprint Backlog* en el tablero el mismo día y se avisa en la daily.

### Pasan al Sprint 3

| Historia | SP | Responsable | Motivo |
|---|---|---|---|
| HU-14 · Cerrar sprint (#46) | 5 | Juan Pablo | Necesita HU-12 y HU-13, que salen al final de este sprint |
| HU-15 · Consultar sprints anteriores (#47) | 2 | Juan Pablo | Necesita HU-14 |
| HU-09 · Priorizar el backlog (#22) | 3 | Mariano | No bloquea a nadie |
| HU-10 · Filtrar y buscar historias (#23) | 2 | Mariano | Prioridad baja |
| HU-17 · Abrir sesión de Planning Poker (#25) | 3 | Mariano | Otra épica (E3): conviene hacerla entera en un sprint |
| HU-18 · Votar de forma oculta (#26) | 5 | Mariano | Ídem |
| HU-19 · Revelar estimaciones (#27) | 3 | Mariano | Ídem |
| TEC-05 · Autenticación de usuarios (#68) | 5 | Mariano | La app todavía no la necesita para la demo; se planifica con la épica de Planning Poker, que sí necesita saber quién vota |
| HU-26 · Consultar y corregir registros de esfuerzo (#4) | 3 | Carolina | Necesita HU-24 |
| HU-28 · Gestionar el estado de un defecto (#6) | 3 | Carolina | Necesita HU-27 |
| HU-29 · Consultar defectos (#7) | 2 | Carolina | Necesita HU-27 |

**Total que pasa: 36 SP.** 33 comprometidos + 36 que pasan = 69 SP.

[A CONFIRMAR] La autenticación (TEC-05) es la que más se discute: es Alta, pero hoy nadie la necesita. Si el equipo
prefiere tenerla antes, entra en lugar de HU-06 y HU-27.

## Decisiones

Las 8 del [refinamiento](refinamiento.md#decisiones-a-confirmar-en-la-planning), con la propuesta:

| # | Decisión | Propuesta | ¿Confirmada? |
|---|---|---|---|
| 1 | ¿Cuándo un sprint está "activo"? | El no cerrado de menor número cuya fecha de inicio ya llegó; uno vencido sin cerrar sigue activo (HU-04 RN2) | Sí (12/10, spec de HU-04 revisada) |
| 2 | Estado del proyecto | Se calcula, no se guarda (HU-04 RN1) | Sí (12/10, spec de HU-04 revisada) |
| 3 | "Hoy" | En la hora de Argentina (HU-04 RN6) | Sí (12/10, spec de HU-04 revisada) |
| 4 | HU-02: fecha de inicio | Tampoco puede quedar después del inicio del primer sprint (HU-02 RN4) | Sí (12/10, spec de HU-02 revisada) |
| 5 | HU-12 y HU-08 | HU-08 define las transiciones en `internal/domain/backlog` y HU-12 las usa | [A CONFIRMAR con Mariano] |
| 6 | HU-13 y HU-08 | HU-08 tiene la regla y HU-13 es la acción desde el sprint activo, que además guarda `completada_en`. Hay que pasar por "En progreso" antes de "Hecho" (CA-08.1) | [A CONFIRMAR con Mariano] |
| 7 | HU-14 y el #80 | El cierre guarda la foto en `sprint_historias` y `sprints.sp_*` | Ya acordado en el #80 (10/10) |
| 8 | HU-12, aviso por velocidad | Con un sprint cerrado, la velocidad es la de ese sprint; sin sprints cerrados no hay aviso | [A CONFIRMAR] |

## Acciones de la Retro del Sprint 1 que se aplican desde hoy

- **Camino crítico primero:** las historias marcadas con orden 1 se empiezan el martes 13/10 y se siguen en cada daily.
- **Revisiones en menos de 24 horas:** la rotación sigue igual (Juan Pablo revisa a Mariano, Mariano a Carolina y
  Carolina a Juan Pablo). Si el revisor no puede en el día, avisa y revisa otro.
- **Lo de producción, el mismo día del merge:** si un PR trae una migración nueva, Juan Pablo la aplica en producción
  ese día. Hoy ninguna historia del sprint la necesita, porque la 0001 ya tiene las 12 tablas.
- **Dailies el mismo día**, de lunes a viernes, en el issue "Sprint 2 · Dailies" (#95).
- **Issue antes de empezar**, y se cierra solo con el merge de su PR (`Closes #n`).

## Riesgos y dependencias

- **Cadena del backlog:** si HU-07, HU-16 o HU-08 se atrasan, HU-12 y HU-13 no llegan. Mientras tanto, Juan Pablo
  avanza con las specs y los escenarios de HU-12 y HU-13.
- **TEC-07 y los conflictos:** cambia `base.html` y los estilos de todas las pantallas. Si llega tarde, cada historia
  que toque plantillas va a tener conflictos. Conviene un PR chico el primer día y después ir mejorando.
- **Sprint más corto:** 6 días en lugar de 7.
- **Las métricas siguen sin pantalla:** el % de SP de HU-04 es la primera métrica que se ve en la app, y va a dar 0 %
  hasta que HU-08 o HU-13 pasen historias a Hecho.

## Dailies

Lunes a viernes, como comentario en el issue "Sprint 2 · Dailies" (#95). Al cerrar el sprint se copian a
[`dailies.md`](dailies.md).
