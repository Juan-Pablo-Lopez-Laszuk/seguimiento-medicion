# Retrospectiva · Sprint 3 · 25/10/2026

Participantes: López, Piastrellini, Bravo · Facilitador: Juan Pablo

> Armada con el tablero, los PRs y las dailies del #115; se discute en la Retro. Lo marcado [A CONFIRMAR] se cierra ahí.

## Métricas del sprint

SP comprometidos: 40 · SP completados: 40 comprometidos y 8 adelantados · Velocidad: 48, la más alta hasta ahora (39 en
el Sprint 1 y 33 en el Sprint 2); el promedio de los tres sprints queda en 40 · Defectos abiertos: 0

| Integrante | Terminado | SP |
|---|---|---|
| Juan Pablo | HU-14, HU-15, HU-37; adelantó HU-38 (25/10) | 12 de 12 (+3) |
| Mariano | HU-16b, TEC-06 (umbral y matriz), TEC-05, HU-09, HU-10 | 15 de 15 |
| Carolina | HU-28, HU-33, HU-34, HU-35; adelantó HU-29 y HU-26 | 13 de 13 (+5) |

## ¿Qué salió bien?

- **Se completó todo lo comprometido y se adelantaron 8 SP:** HU-29 y HU-26 (Carolina) y HU-38 (Juan Pablo). Carolina
  terminó todas sus historias el miércoles 21/10.
- **El camino crítico entró el primer día:** HU-14 (cerrar sprint) se mergeó el lunes 19/10, y con ella pudieron avanzar
  HU-15, HU-35 y HU-37.
- **TEC-05 se adelantó y quedó en producción:** el login con Supabase Auth se mergeó el 23/10 y ese mismo día quedó
  configurado en Supabase y Vercel. Planning Poker ya puede saber quién vota.
- **Lo de producción se hizo el día del merge:** la migración 0003 de HU-16b se aplicó antes de mergear #414, como pidió
  Mariano, y las variables de TEC-05 se cargaron el mismo día. No hubo errores en producción.
- **La calidad quedó controlada por el CI:** el umbral de cobertura del 80 % (TEC-06) y la matriz de trazabilidad con su
  test (51 de 59 requerimientos cubiertos). Los escenarios BDD pasaron de 221 a 384 en verde.
- **TDD verificable en todas las historias:** cada revisión comprobó commit por commit que los RED fallan y los GREEN
  pasan, y probó la historia en la app local, también con PostgreSQL.
- **Revisiones rápidas casi siempre:** 14 de los 18 PRs del sprint se aprobaron en menos de 24 horas; HU-33 (#116),
  HU-14 (#117) y HU-15 (#413) en menos de una hora.
- **Las tareas como sub-issues con horas**, como pidieron los profesores: las 48 HU y TEC quedaron divididas el 20/10.
- **Lo que no iba a llegar se avisó a tiempo:** Mariano avisó el 22/10 que, si no llegaba, HU-10 pasaba al Sprint 4.
  Finalmente llegó.

## ¿Qué podemos mejorar?

- **Las revisiones de Juan Pablo se acumularon del 21 al 23/10:** Mariano tuvo hasta cuatro PRs esperándolo a la vez.
  - TEC-05 (#418) esperó unas 46 horas.
  - HU-09 (#421) y HU-10 (#422), unas 37 y 36.
  - La matriz de TEC-06 (#419) y el arreglo de HU-06 (#420), unas 23.
  - Fuera de las de Juan Pablo, solo HU-29 (#412) pasó las 24 horas (unas 34), por los conflictos con `main`.
- **Las dailies del Agile Enabler:** Juan Pablo dejó la del 20/10 y después una sola el 24/10 con las del 21 al 24.
  Justo era quien tenía que revisar el issue cada noche (acción del Sprint 2).
- **El viernes 23 y el sábado 24 Carolina y Mariano no dejaron daily.** Mariano cargó la del 21/10 recién el 22/10.
- **Un PR apilado se cerró solo:** al borrarse la rama de HU-34, GitHub le cambió la base a #412 y lo cerró. Mariano
  lo reabrió.
- **Conflictos repetidos en los mismos archivos:** `docs/ia/registro.md` (cada PR suma una fila), `router.go`,
  `app.go` y la matriz de trazabilidad. Hubo que resolverlos en #119, #120 y #412 (dos veces), y HU-09 (#421) tuvo que integrarse con el login después del
  merge de TEC-05 (#418).
- **Otra vez trabajo que entra el último día:** HU-38 (#423) se abrió el sábado 24/10 y se mergeó el domingo 25/10.
  Salió porque era adelantada, pero no quedaba margen para observaciones.
- **Las horas reales del equipo no se registraron:** las tareas tienen horas estimadas pero no las trabajadas, así que
  no se puede calcular la desviación del equipo (lo dice el informe de métricas, #427).
- **Un PR aprobado quedó sin mergear:** la Planning del Sprint 4 (#425) se aprobó el 25/10 y siguió abierta.

## Acciones

El Sprint 4 tiene 4 días de código: el jueves 29/10 se congela.

| Acción | Responsable | Para cuándo |
|---|---|---|
| Con dos o más PRs esperando su revisión, Juan Pablo revisa antes de seguir con su propio código. Ninguna revisión pasa las 24 horas | Juan Pablo | Sprint 4 |
| Daily el mismo día, también la del Agile Enabler, que la deja antes de revisar las de los demás | Todos · controla Juan Pablo | Sprint 4 |
| Todo lo que entra al congelamiento está en revisión el miércoles 28/10 a la noche | Todos | 28/10 |
| Antes de mergear la base de un PR apilado, se le cambia la base a `main` al PR que depende de ella, así GitHub no lo cierra | El autor del PR apilado | Sprint 4 |
| El autor del PR que se mergea primero avisa en la daily, y los demás traen `main` ese mismo día | Todos | Sprint 4 |
| Al cerrar cada tarea del Sprint 4, se anotan en ella las horas que llevó [A CONFIRMAR] | Todos | Sprint 4 |
| Un PR aprobado lo mergea el revisor en el momento, salvo que el autor pida esperar | Revisor | Sprint 4 |

## Seguimiento de las acciones del sprint anterior

| Acción del Sprint 2 | ¿Se cumplió? |
|---|---|
| Revisión en menos de 24 horas. Si el revisor no puede conectarse, lo avisa y revisa el siguiente de la rotación | Parcial: 14 de 18 PRs a tiempo; #418, #421 y #422 (Juan Pablo) y #412 tardaron entre 34 y 46 horas |
| Las consultas que esperan una decisión se contestan en el día | Sí: las decisiones de la Planning (#114) y las de TEC-05 se confirmaron antes de programar [A CONFIRMAR] |
| En la Planning, cada dependencia se verifica contra el código que ya está en `main` | Sí: HU-14 fue primera y no hubo dependencias falsas |
| Después de cada merge, los PRs abiertos traen `main` ese mismo día | Sí, aunque los conflictos en los mismos archivos siguieron (ver arriba) |
| Lo que no esté en revisión el viernes se avisa en la daily | Sí: Mariano avisó por HU-10, que finalmente llegó |
| Daily el mismo día; Juan Pablo revisa el issue cada noche | No: Juan Pablo cargó las del 21 al 24/10 juntas y nadie dejó daily el 23 y el 24 |
| HU-14 va primera en el Sprint 3 | Sí: se mergeó el 19/10, el primer día |
