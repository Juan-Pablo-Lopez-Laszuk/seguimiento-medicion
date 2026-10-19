# Retrospectiva · Sprint 2 · 18/10/2026

Participantes: López, Piastrellini, Bravo · Facilitador: Juan Pablo

> Armada con el tablero, los PRs y las dailies del #95; se discute en la Retro.

## Métricas del sprint

SP comprometidos: 33 · SP completados: 33 (27 al sábado 17/10 y 6 el domingo 18/10) · Velocidad: 33; el promedio de
los dos sprints de construcción queda en 36. El sprint tuvo 6 días en lugar de 7 · Defectos abiertos: 0

| Integrante | Terminado | SP |
|---|---|---|
| Juan Pablo | HU-04, HU-02, HU-12, HU-13 | 12 de 12 |
| Mariano | HU-07, HU-08, HU-16 (sin el historial), HU-06 | 9 de 9 |
| Carolina | TEC-07, HU-25, HU-24 (18/10), HU-27 (18/10) | 12 de 12 |

## ¿Qué salió bien?

- **La Planning se hizo el lunes** (12/10) con el compromiso ya cargado en el tablero, y las decisiones que esperaban a
  Mariano (5, 6 y 8) quedaron confirmadas por escrito en el #95 el primer día, antes de programar HU-12 y HU-13.
- **El camino crítico se atacó primero:** HU-07 estuvo en `main` el 13/10 y HU-08 el 15/10 a la mañana. Juan Pablo
  terminó HU-12 y HU-13 ese mismo día.
- **Se destrabó HU-16 dividiéndola:** el historial de estimaciones necesitaba una migración nueva; pasó a HU-16b (#106)
  y HU-16 salió sin migración.
- **Las revisiones de este sprint ya no dependen de acordarse:** el workflow de #104 pide la revisión según la rotación
  apenas se abre un PR. Funcionó desde #107.
- **TDD verificable:** todas las historias tienen sus commits RED → GREEN, y las revisiones los comprobaron commit por
  commit (#102, #105, #107). Los escenarios BDD pasaron de 154 (HU-08) a 221 en verde.
- **PRs apilados** cuando una historia dependía de otra todavía en revisión (#105 → #107, #110 → #111): nadie tuvo que
  esperar el merge para seguir.
- **Sin cambios en producción:** ninguna historia del sprint necesitó migración (todas usan el esquema de TEC-03), así
  que no hubo pasos manuales en Supabase ni en Vercel.

## ¿Qué podemos mejorar?

- **Revisiones que tardaron más de 24 horas:**
  - HU-12 (#105) esperó unas 55 horas y HU-13 (#107) unas 46: Carolina no pudo conectarse del 15 al 17/10 y nadie más
    las tomó.
  - HU-16 (#108) esperó unas 47 horas y HU-06 (#109) unas 30, las dos de Juan Pablo.
  - Las del principio del sprint sí se hicieron a tiempo: HU-07 (#100) en 5 horas y HU-08 (#103) en 18.
- **Una decisión frenó dos días:** Mariano consultó por el historial de HU-16 el 13/10 y la decisión llegó el 15/10.
  No lo dejó parado porque siguió con HU-08, pero HU-16 recién entró el último día.
- **Una dependencia de la Planning no era real:** HU-12 figuraba como dependiente de HU-16, pero solo necesita saber si
  la historia tiene story points, y eso ya existe desde HU-05.
- **Conflictos en cada merge:** todas las historias agregan líneas en los mismos archivos (`app/app.go`, `router.go`,
  `router_test.go`, `features_test.go`, los repositorios de historias y `docs/ia/registro.md`). Hubo que resolver
  conflictos en #104/#105, en #108/#109 y ahora en #110.
- **Un merge sin revisión terminada:** #104 se mergeó por error antes de terminar de revisarlo.
- **Trabajo que llega a revisión el último día:** HU-24 (#110) y HU-27 (#111) se abrieron el sábado 17/10 a la noche y
  se mergearon el domingo 18/10 a la noche, horas antes del cierre. La revisión de Mariano fue en menos de 24 horas,
  pero no quedaba margen: con cualquier observación, las dos pasaban al Sprint 3.
- **Las dailies siguen atrasadas**, la misma acción del Sprint 1 que no se cumplió:
  - Juan Pablo cargó las del 14 al 17/10 juntas el 17/10.
  - Mariano cargó las del 13 y 14/10 el 14/10, y las del 15 y 16/10 el 16/10.
  - Carolina cargó las del 15 al 17/10 juntas el 17/10.
  - El domingo 18/10 nadie cargó la daily.
  - El atraso de #105 y #107 no se vio a tiempo porque no había dailies que lo mostraran.

## Acciones

| Acción | Responsable | Para cuándo |
|---|---|---|
| Revisión en menos de 24 horas. Si el revisor no puede conectarse, lo avisa en el issue del sprint y revisa el siguiente de la rotación | Todos | Sprint 3 |
| Las consultas que esperan una decisión (Agile Enabler o equipo) se contestan en el día, aunque sea con "lo decido mañana a la mañana" | Juan Pablo | Sprint 3 |
| En la Planning, cada dependencia se verifica contra el código que ya está en `main` antes de ordenar las historias | Juan Pablo | Planning del Sprint 3 |
| Después de cada merge, los PRs abiertos traen `main` ese mismo día (lo hace el autor), así los conflictos quedan chicos | Todos | Sprint 3 |
| Lo que no esté en revisión el viernes se avisa en la daily y se decide si entra o pasa al sprint siguiente | Todos · controla Juan Pablo | Sprint 3 |
| Daily el mismo día. Juan Pablo revisa el issue cada noche y le avisa a quien falte | Todos · controla Juan Pablo | Sprint 3 |
| HU-14 (cerrar sprint) va primera en el Sprint 3: sin ella la velocidad no se calcula desde la app | Juan Pablo | Sprint 3 |

## Seguimiento de las acciones del sprint anterior

| Acción del Sprint 1 | ¿Se cumplió? |
|---|---|
| La Planning se hace el lunes; el domingo se dejan listos los datos y el refinamiento | Sí: Planning el lunes 12/10, con el refinamiento del 09/10 |
| Lo que bloquea a otros se marca en la Planning, se empieza primero y se revisa en cada daily | Sí: la Planning ordenó el camino crítico y HU-07 y HU-08 fueron lo primero que se revisó y mergeó; HU-16 se demoró por la decisión del historial |
| Una revisión pedida se hace en menos de 24 horas | Parcial: se cumplió con #100, #103, #110 y #111, pero #105, #107, #108 y #109 tardaron entre 30 y 55 horas |
| Lo que hay que hacer en producción se hace el mismo día que se mergea el PR | No aplicó: ninguna historia del sprint trajo migración |
| Daily de lunes a viernes, el mismo día, también del Agile Enabler | No: los tres cargamos dailies atrasadas (ver arriba) |
| Toda tarea tiene su issue antes de empezarla | Parcial: HU-16b (#106) se creó antes de empezarla, pero el workflow de #104 se hizo sin issue |
