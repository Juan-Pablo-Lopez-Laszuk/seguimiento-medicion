# Sprint Planning · Sprint 3 · 19/10/2026

Participantes: López, Piastrellini, Bravo · Facilitador: Juan Pablo

> Armada el 19/10 a partir del tablero, la [Review](../sprint-2/review.md) y la [Retro del Sprint 2](../sprint-2/retro.md).
> Lo marcado [A CONFIRMAR] lo cierra el equipo comentando en el PR de esta acta.

## Fechas

Del **lunes 19/10 al domingo 25/10** (7 días). El Sprint 4 (26/10 al 01/11) es el de cierre: exportación a PDF,
documentación y entrega final.

## Sprint Goal

> Métricas completas y a la vista: cerrar sprints con su foto, ver los sprints anteriores, las horas y los defectos
> por sprint, el panel de indicadores del proyecto y el reporte de un sprint; y sostener la calidad con pruebas de
> integración y un umbral de cobertura en el CI.

[A CONFIRMAR]

## Capacidad y velocidad

- Velocidad: **39 SP** en el Sprint 1 (8 días) y **33 SP** en el Sprint 2 (6 días). El promedio es **36 SP**. Este
  sprint tiene 7 días y los tres están disponibles toda la semana.
- El tablero tiene **72 SP** marcados para el Sprint 3 (67 del Sprint 3, más HU-16b y HU-28 que vienen del Sprint 2).
  Se proponen **35 SP**, unos 12 por persona.
- Criterio del recorte:
  1. Primero lo que destraba a otros (acción de la Retro del Sprint 1, que funcionó en el Sprint 2). **HU-14** es lo
     único que falta para que la velocidad salga de la app, y la usan HU-15, HU-35 y HU-37. HU-28 y HU-33 dan los
     datos de defectos y horas que muestran HU-34, HU-35 y HU-37.
  2. Después, lo que pide el objetivo del sprint en el README: métricas completas (HU-33 y HU-34), panel (HU-35),
     reportes (HU-37) y robustez (TEC-06).
  3. Mariano trabaja backlog y calidad (HU-09, HU-10, HU-16b y TEC-06). La épica de Planning Poker (HU-17 a HU-22)
     y la autenticación (TEC-05) pasan enteras al Sprint 4, para hacerlas juntas (decisión de Juan Pablo del 19/10).

## Historias del sprint

### Comprometidas

| Orden | Historia | SP | Responsable | Depende de | Por qué en este orden |
|---|---|---|---|---|---|
| 1 | HU-14 · Cerrar sprint (#46) | 5 | Juan Pablo | HU-12, HU-13 (hechas) | Camino crítico: sin sprints cerrados no hay velocidad, ni sprints anteriores, ni reporte |
| 1 | HU-28 · Gestionar el estado de un defecto (#6) | 3 | Carolina | HU-27 (hecha) | Ya está en revisión (#112). HU-34 necesita los defectos resueltos |
| 1 | HU-33 · Horas estimadas, reales y desviación (#11) | 3 | Carolina | HU-23, HU-24 (hechas) | La desviación la muestran HU-35 y HU-37 |
| 1 | HU-16b · Historial de estimaciones (#106) | 2 | Mariano | HU-16 (hecha) | Trae la migración 0003: mejor temprano, para aplicarla en producción con tiempo |
| 1 | TEC-06 · Pruebas de integración y umbral de cobertura (#69) | 3 | Mariano | — | Cuanto antes esté el umbral en el CI, más historias del sprint lo cumplen desde el principio |
| 2 | HU-15 · Consultar sprints anteriores (#47) | 2 | Juan Pablo | HU-14 | Usa la foto del cierre |
| 2 | HU-34 · Defectos detectados y resueltos (#12) | 2 | Carolina | HU-27 (hecha), HU-28 | Necesita el sprint de resolución de HU-28 |
| 2 | HU-09 · Priorizar el backlog (#22) | 3 | Mariano | HU-05 (hecha) | No bloquea a nadie |
| 3 | HU-37 · Reporte de sprint (#48) | 5 | Juan Pablo | HU-14, HU-33, HU-34 | Junta las métricas de un sprint cerrado |
| 3 | HU-35 · Panel de indicadores del proyecto (#13) | 5 | Carolina | HU-14, HU-33, HU-28 | Muestra velocidad, desviación y defectos abiertos |
| 3 | HU-10 · Filtrar y buscar historias (#23) | 2 | Mariano | HU-05 (hecha) | Prioridad baja |

**Total comprometido: 35 SP** · Juan Pablo 12 · Mariano 10 · Carolina 13 (3 de ellos ya en revisión).

Las dependencias se verificaron contra el código que ya está en `main` (acción de la Retro del Sprint 2): HU-12,
HU-13, HU-23, HU-24, HU-27 y las métricas de HU-30, HU-31 y HU-32 están mergeadas.

### Si sobra tiempo (no comprometidas)

| Historia | SP | Responsable | Condición |
|---|---|---|---|
| HU-17 · Abrir sesión de Planning Poker (#25) | 3 | Mariano | Si sus cuatro historias están en `main`; adelanta la épica del Sprint 4 |
| HU-29 · Consultar defectos (#7) | 2 | Carolina | Después de HU-28 |
| HU-38 · Reporte de proyecto (#49) | 3 | Juan Pablo | Después de HU-37 |
| HU-26 · Consultar y corregir registros de esfuerzo (#4) | 3 | Carolina | — |
| TEC-07b · Componentes de formularios y tablas (#101) | 2 | Carolina | — |

Si se empieza alguna, se pasa a *Sprint Backlog* en el tablero el mismo día y se avisa en la daily.

### Pasan al Sprint 4

| Historia | SP | Responsable | Motivo |
|---|---|---|---|
| HU-17 · Abrir sesión de Planning Poker (#25) | 3 | Mariano | La épica de Planning Poker va entera en el Sprint 4, junto con la autenticación |
| HU-18 · Votar de forma oculta (#26) | 5 | Mariano | Ídem |
| HU-19 · Revelar estimaciones (#27) | 3 | Mariano | Ídem |
| HU-20 · Detectar diferencias entre estimaciones (#28) | 3 | Mariano | Ídem |
| HU-21 · Realizar nuevas rondas (#29) | 3 | Mariano | Ídem |
| HU-22 · Registrar la estimación acordada (#30) | 2 | Mariano | Ídem |
| TEC-05 · Autenticación de usuarios (#68) | 5 | Mariano | La necesita Planning Poker para saber quién vota |
| HU-36 · Gráficos de métricas (#14) | 5 | Carolina | Va después del panel (HU-35) |
| HU-38 · Reporte de proyecto (#49) | 3 | Juan Pablo | Si no entra como "si sobra tiempo" |
| HU-26 · Consultar y corregir registros de esfuerzo (#4) | 3 | Carolina | Ídem |
| HU-29 · Consultar defectos (#7) | 2 | Carolina | Ídem |

**Total que pasa: 37 SP.** 35 comprometidos + 37 que pasan = 72 SP.

## Riesgo principal: el Sprint 4

Con la épica de Planning Poker y la autenticación (24 SP), HU-36 (5), HU-39 (exportar a PDF, 5) y la documentación
de la entrega, el **Sprint 4 ya arranca con al menos 34 SP** y es el sprint de cierre. Para que no se desborde:

- Si Mariano termina sus historias, adelanta HU-17 en este sprint.
- En la Planning del Sprint 4 se decide qué parte de Planning Poker es imprescindible para la entrega.

[A CONFIRMAR]

## Decisiones

| # | Decisión | Propuesta | ¿Confirmada? |
|---|---|---|---|
| 1 | HU-14: qué pasa con las historias no terminadas | Vuelven al backlog como Pendientes y sin sprint (CA-14.2), pero quedan en la foto del cierre (`sprint_historias`), así el sprint conserva lo que tenía planificado (#80) | [A CONFIRMAR] |
| 2 | HU-14: `EnSprintCerrado` del repositorio en memoria | HU-14 lo actualiza para que consulte el sprint de la historia y la foto del cierre, como el de PostgreSQL (hoy devuelve siempre `false`) | Sí (Review del Sprint 2) |
| 3 | HU-16b: migración 0003 | Juan Pablo la aplica en producción el mismo día del merge. El `CHECK` de fechas de `sprints` va en la 0004, en otro PR (#106) | [A CONFIRMAR con Mariano] |
| 4 | TEC-06: umbral de cobertura | El CI falla si la cobertura combinada de `internal/domain` e `internal/metrics` baja del 80 % (meta interna, #69) | [A CONFIRMAR con Mariano] |
| 5 | HU-35: un indicador sin datos | Se muestra la tarjeta vacía con "sin datos" en lugar de 0 (por ejemplo, la velocidad sin sprints cerrados) (CA-35.2) | [A CONFIRMAR con Carolina] |

## Acciones de la Retro del Sprint 2 que se aplican desde hoy

- **Revisiones en menos de 24 horas.** La rotación sigue igual: Juan Pablo revisa a Mariano, Mariano a Carolina y
  Carolina a Juan Pablo, y el workflow de #104 pide la revisión solo. Si el revisor no puede conectarse, lo avisa en el
  issue del sprint y revisa el siguiente de la rotación.
- **Las consultas que esperan una decisión se contestan en el día.**
- **Después de cada merge, los PRs abiertos traen `main` ese mismo día** (lo hace el autor).
- **Lo que no esté en revisión el viernes 23/10 se avisa en la daily** y se decide si entra o pasa al Sprint 4.
- **Daily el mismo día**, en el issue "Sprint 3 · Dailies" (#115). Juan Pablo revisa el issue cada noche y le avisa a quien
  falte.
- **HU-14 va primera.**

## Riesgos y dependencias

- **HU-14 es el camino crítico:** si se atrasa, HU-15, HU-35 y HU-37 no tienen sprints cerrados con los que trabajar.
  Juan Pablo la empieza hoy, y mientras tanto los demás avanzan con sus specs y escenarios.
- **HU-28 tiene el CI en rojo** (#112). Hay que arreglarlo antes de la revisión, porque de ella dependen HU-34 y HU-35.
- **Migración 0003:** es la primera migración desde TEC-03. Hay que aplicarla en producción el mismo día del merge, o
  el deploy de Vercel queda con un esquema viejo.
- **Conflictos:** HU-14, HU-35 y HU-37 tocan las mismas pantallas y el mismo `app.go`. Se resuelven trayendo `main` el
  mismo día de cada merge.

## Dailies

Lunes a viernes, como comentario en el issue "Sprint 3 · Dailies" (#115). Al cerrar el sprint se copian a
[`dailies.md`](dailies.md).
