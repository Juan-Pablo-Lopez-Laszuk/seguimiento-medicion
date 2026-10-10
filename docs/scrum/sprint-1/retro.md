# Retrospectiva · Sprint 1 · 11/10/2026

Participantes: López, Piastrellini, Bravo · Facilitador: Juan Pablo

> **Borrador armado el viernes 09/10**: los puntos son una propuesta a partir de lo que pasó en el repo y en el
> tablero. Cada uno agrega, saca o corrige en la reunión del domingo.

## Métricas del sprint

SP comprometidos: 39 (19 ya terminados al hacer la Planning + 20 por hacer) · SP completados al 09/10: 29 · Velocidad:
[A COMPLETAR el domingo] · Defectos abiertos: 0

| Integrante | Terminado | SP |
|---|---|---|
| Juan Pablo | TEC-01, TEC-02, HU-01, HU-11, HU-03 | 14 |
| Carolina | #71, HU-30, HU-31, HU-32, HU-23 | 12 |
| Mariano | #83, #84 | 3 |
| [A COMPLETAR] | TEC-03, HU-05, TEC-04, #80 | hasta 10 |

## ¿Qué salió bien?

- **TDD visible en el historial:** todas las historias tienen commits `test(...): RED` antes de `feat(...): GREEN`, y
  los revisores comprobaron que cada RED falle (#76, #79, #88).
- **La revisión cruzada funcionó y se probó en el preview:** Carolina aprobó HU-11 después de probar cada caso en el
  preview de Vercel (#88).
- **Las specs se revisaron antes del código:** las decisiones de HU-11 y HU-03 se acordaron antes de programar y
  quedaron explicadas en el PR.
- **El CI frenó problemas antes de `main`:** formato, linter y tests en cada PR.
- **PRs chicos y seguidos:** se mergearon 16 PRs en la semana.

## ¿Qué podemos mejorar?

- **La Planning se hizo tarde** (el miércoles 07/10 en lugar del lunes 05/10): el sprint arrancó con dos días sin
  objetivo acordado.
- **El camino crítico arrancó tarde:** TEC-03 era lo que más bloqueaba (TEC-04 y el deploy real dependen de él) y llegó
  a revisión el viernes 09/10 con el CI en rojo (se corrigió el 10/10). Mientras tanto, los datos de la app se guardan en memoria y en Vercel
  se pierden.
- **Dailies:** en el issue #86 Carolina las dejó el día que correspondía (08 y 09/10); Mariano cargó las del 05 al
  07/10 recién el 09/10, y Juan Pablo no dejó ninguna. [A COMPLETAR]
- **Estado del tablero distinto de `main`:** #1 se cerró a mano antes de mergear su PR (#89), y el tablero contó 2 SP
  de más durante unas horas. Lo que cierra una issue es el merge del PR (`Closes #n`).
- **Issues creadas después del trabajo:** #83 y #84 se crearon y cerraron el mismo día, al terminar. [A COMPLETAR si
  se repitió]
- **Revisiones que tardaron:** HU-32 (#87) esperó un día la revisión. [A COMPLETAR]

## Acciones

| Acción | Responsable | Para cuándo |
|---|---|---|
| La Planning se hace el lunes a primera hora; el domingo se dejan listos los datos y el refinamiento | Juan Pablo | Lunes 12/10 |
| Lo que bloquea a otros (camino crítico) se marca en la Planning y se revisa en cada daily | Todos · controla Juan Pablo | Sprint 2 |
| Una issue se cierra solo con el merge de su PR (`Closes #n`), nunca a mano | Todos | Sprint 2 |
| Daily de lunes a viernes en el issue del sprint, también del Agile Enabler | Todos | Sprint 2 |
| Una revisión pedida se hace en menos de 24 horas | Todos | Sprint 2 |
| [A COMPLETAR] | | |

## Seguimiento de las acciones del sprint anterior

| Acción del Sprint 0 | ¿Se cumplió? |
|---|---|
| Mergear TEC-01 y armar el CI (TEC-02) antes de empezar las historias | Sí: los dos el lunes 05/10 (#70 y #75) |
| Todo issue del sprint lleva SP antes de la Planning | Sí: todas las historias del sprint tenían SP al cerrar la Planning |
| Un PR no se aprueba sin el preview de Vercel en verde | Sí: los PRs mergeados tenían el preview en verde; en HU-11 (#88) la revisión documentó la prueba caso por caso |
| Daily por mensaje en el issue del sprint | Parcial: Carolina al día, Mariano cargó las atrasadas el 09/10 y Juan Pablo ninguna (#86) |
