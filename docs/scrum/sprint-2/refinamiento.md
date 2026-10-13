# Refinamiento · Sprint 2 · 09/10/2026

> Preparación de la Planning del lunes 12/10. Revisa que las historias del Sprint 2 cumplan la Definition of Ready
> ([`../definition-of-ready-done.md`](../definition-of-ready-done.md)) y deja anotadas las dependencias y las decisiones
> que hay que tomar **antes** de programar. Lo armó Juan Pablo; las decisiones marcadas "a confirmar" se cierran en la
> Planning.

## Definition of Ready

Las 23 historias que hoy tienen Sprint = 2 en el tablero están en formato Como / Quiero / Para, tienen criterios de
aceptación, story points y responsable. Lo que falta es **marcar dependencias** y **recortar**: suman **69 SP** y el
Sprint 1 terminó cerca de 30 (ver la Retro del Sprint 1), así que en la Planning hay que elegir.

| Responsable | Historias | SP |
|---|---|---|
| Juan Pablo | HU-02, HU-04, HU-12, HU-13, HU-14, HU-15 | 19 |
| Mariano | HU-06, HU-07, HU-08, HU-09, HU-10, HU-16, HU-17, HU-18, HU-19, TEC-05 | 30 |
| Carolina | HU-24, HU-25, HU-26, HU-27, HU-28, HU-29, TEC-07 | 20 |

HU-05 y TEC-04 no se suman: el Sprint 1 se estiró hasta el lunes 12/10 y se cerraron ahí (ver la Review del Sprint 1).

## Cadena de dependencias

La mayor parte del Sprint 2 cuelga de **las historias del backlog** (E2, Mariano). Sin ellas no se puede asignar, ni
completar, ni cerrar un sprint:

```
TEC-03 (base) ──► HU-05 (crear historia) ──► HU-07 (criterios) ──► HU-08 (estados) ──► HU-13 (completar)
                                       ├──► HU-16 (estimar) ─────────────────────────► HU-12 (asignar al sprint)
                                       └──► HU-25 (tareas, Carolina) ─► HU-24 (esfuerzo) ─► HU-26
HU-12 + HU-13 ──► HU-14 (cerrar sprint) ──► HU-15 (sprints anteriores)
HU-03 (hecha) ──► HU-24 (esfuerzo de un integrante)
```

**Propuesta para la Planning:** TEC-03 y HU-05 ya están en `main`, así que la cadena arranca desbloqueada. Mariano
empieza por HU-07, HU-08 y HU-16, que destraban al resto; mientras tanto Juan Pablo hace HU-04 y HU-02 y Carolina
TEC-07, HU-25 y HU-27 a HU-29.

## Historias de Juan Pablo

| Historia | SP | Depende de | Estado del refinamiento |
|---|---|---|---|
| HU-04 · Consultar estado del proyecto (#42) | 3 | HU-01, HU-03, HU-11 (hechas); HU-05 solo para el % de SP | **Spec y escenarios en borrador** (`specs/HU-04-estado-proyecto.md`) |
| HU-02 · Modificar proyecto (#40) | 2 | HU-04 (estado Finalizado) | **Spec y escenarios en borrador** (`specs/HU-02-modificar-proyecto.md`) |
| HU-12 · Asignar historias a un sprint (#44) | 5 | HU-05, HU-07, HU-16 | Preguntas abiertas (abajo) |
| HU-13 · Registrar historias completadas (#45) | 2 | HU-12, HU-08 | Preguntas abiertas (abajo) |
| HU-14 · Cerrar sprint (#46) | 5 | HU-12, HU-13 | Preguntas abiertas (abajo) |
| HU-15 · Consultar sprints anteriores (#47) | 2 | HU-14 | Sin dudas: usa la foto del cierre (`sprint_historias`) y HU-32 |

## Decisiones a confirmar en la Planning

1. **¿Cuándo un sprint está "activo"?** Ninguna historia lo dice, pero HU-04, HU-13 y HU-14 lo necesitan.
   Propuesta (HU-04 RN2): es el sprint **no cerrado de menor número cuya fecha de inicio ya llegó**. No hay un botón
   "Iniciar sprint": se activa solo por fecha y deja de estarlo al cerrarlo (HU-14). Un sprint vencido sin cerrar sigue
   activo, así se puede cerrar tarde. Afecta a la métrica de velocidad de Carolina (`metrics.SprintActivo`), que hoy
   recibe el estado ya resuelto.
2. **El estado del proyecto se calcula, no se guarda** (HU-04 RN1). La columna `proyectos.estado` de TEC-03 queda como
   está y se puede quitar más adelante.
3. **"Hoy" en la hora de Argentina** (HU-04 RN6). Vercel corre en UTC; de 21 a 24 h el día sería el siguiente.
4. **HU-02:** ¿la fecha de inicio del proyecto tampoco puede quedar después del inicio del primer sprint? La historia
   solo lo dice para la fecha de fin (CA-02.2), pero sin esto un sprint quedaría fuera del proyecto. Propuesta: sí
   (HU-02 RN4).
5. **HU-12 y HU-08 (Mariano):** al asignar una historia a un sprint pasa a "En sprint" (CA-08.1). ¿La transición la
   hace HU-12 llamando a la regla de HU-08, o HU-08 expone una función que HU-12 usa? Propuesta: HU-08 define las
   transiciones en `internal/domain/backlog` y HU-12 las usa, sin duplicarlas.
6. **HU-13 y HU-08 se pisan:** las dos pasan una historia a Hecho. Propuesta: HU-08 tiene la regla (transiciones y
   criterios cumplidos, CA-08.2) y HU-13 es la acción desde la vista del sprint activo, que además guarda la fecha
   (`completada_en`). Hay que acordar si se puede pasar de "En sprint" a "Hecho" directo o hay que pasar por "En
   progreso" (CA-08.1 dice que hay que pasar).
7. **HU-14 y el #80:** el cierre guarda la foto en `sprint_historias` y en `sprints.sp_planificados` /
   `sp_completados`, y las métricas de los sprints cerrados se leen de ahí (RG-6 de `docs/metricas.md`). Es la
   propuesta del #80, que sigue sin confirmar con Mariano.
8. **HU-12, advertencia por velocidad (CA-12.3):** con un solo sprint cerrado la velocidad es la de ese sprint (HU-31);
   sin sprints cerrados no hay advertencia.

## Para avisar a Mariano

- HU-02 necesita en el repositorio de proyectos `Actualizar` y buscar nombre repetido **sin contar al propio proyecto**.
  No entró en TEC-03, así que se agrega con HU-02.
- HU-04 lee historias por proyecto (para el % de SP): lo cubre el repositorio de HU-05, ya mergeado.
