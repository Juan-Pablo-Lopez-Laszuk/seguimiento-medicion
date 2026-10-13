# Retrospectiva · Sprint 1 · 12/10/2026

Participantes: López, Piastrellini, Bravo · Facilitador: Juan Pablo

## Métricas del sprint

SP comprometidos: 39 (19 ya terminados al hacer la Planning + 20 por hacer) · SP completados: 39 (33 al domingo 11/10
y 6 el lunes 12/10, con el sprint estirado un día) · Velocidad: 39 (primer sprint de construcción; el Sprint 0 fue de
preparación y no se promedia) · Defectos abiertos: 0

| Integrante | Terminado | SP |
|---|---|---|
| Juan Pablo | TEC-01, TEC-02, HU-01, HU-11, HU-03 | 14 |
| Carolina | #71, HU-30, HU-31, HU-32, HU-23, TEC-04 (12/10) | 15 |
| Mariano | #83, #84, TEC-03, HU-05 (12/10) | 9 |
| Juan Pablo y Mariano | #80 (decisión) | 1 |

## ¿Qué salió bien?

- **TDD visible en el historial:** todas las historias tienen commits `test(...): RED` antes de `feat(...): GREEN`, y
  los revisores comprobaron que cada RED falle (#76, #79, #88).
- **La revisión cruzada funcionó y se probó en el preview:** Carolina aprobó HU-11 después de probar cada caso en el
  preview de Vercel (#88).
- **Las revisiones encontraron problemas reales:** en TEC-03 (#91) la revisión contra una base real detectó que Vercel
  no subía `migrations/` y cinco errores del linter, y se corrigieron antes de mergear.
- **Las specs se revisaron antes del código:** las decisiones de HU-11 y HU-03 se acordaron antes de programar y
  quedaron explicadas en el PR.
- **El CI frenó problemas antes de `main`:** formato, linter y tests en cada PR.
- **PRs chicos y seguidos:** se mergearon 16 PRs en la semana.

## ¿Qué podemos mejorar?

- **La Planning se hizo tarde** (el miércoles 07/10 en lugar del lunes 05/10): el sprint arrancó con dos días sin
  objetivo acordado.
- **El camino crítico arrancó tarde:** TEC-03 era lo que más bloqueaba (HU-05, TEC-04 y el deploy real dependen de él)
  y llegó a revisión el viernes 09/10 con el CI en rojo (se corrigió el 10/10). HU-05 recién se pudo empezar el sábado
  10/10 y no llegó al domingo: hubo que estirar el sprint un día.
- **Revisiones que tardaron:** HU-05 (#93) esperó dos días la revisión de Juan Pablo (del sábado 10/10 a la noche al
  lunes 12/10 a la noche), y fue una de las causas de estirar el sprint. HU-32 (#87) esperó casi un día.
- **TEC-04 esperó la configuración de producción:** el código se mergeó el sábado 10/10, pero la base de producción
  recién se cargó en Vercel el lunes 12/10. Las credenciales de producción las maneja solo Juan Pablo, así que ese
  paso dependía de una sola persona.
- **Dailies:** en el issue #86 Carolina dejó las del 08 y el 09/10; Mariano cargó las del 05 al 07/10 recién el 09/10
  y después las del 09 y el 10/10; Juan Pablo cargó todas las suyas juntas el 12/10. Las dailies atrasadas no sirven
  para detectar bloqueos a tiempo.
- **Estado del tablero distinto de `main`:** #1 se cerró a mano antes de mergear su PR (#89), y el tablero contó 2 SP
  de más durante unas horas. Lo que cierra una issue es el merge del PR (`Closes #n`).
- **Issues creadas después del trabajo:** #83 y #84 se crearon y cerraron el mismo día, al terminar.

## Acciones

| Acción | Responsable | Para cuándo |
|---|---|---|
| La Planning se hace el lunes; el domingo se dejan listos los datos y el refinamiento | Juan Pablo | Cada sprint |
| Lo que bloquea a otros (camino crítico) se marca en la Planning, se empieza primero y se revisa en cada daily | Todos · controla Juan Pablo | Sprint 2 |
| Una revisión pedida se hace en menos de 24 horas; si el revisor no puede, avisa y revisa otro | Todos | Sprint 2 |
| Lo que hay que hacer en producción (migraciones, variables en Vercel) se hace el mismo día que se mergea el PR que lo necesita | Juan Pablo | Sprint 2 |
| Daily de lunes a viernes en el issue del sprint, el mismo día, también del Agile Enabler | Todos | Sprint 2 |
| Una issue se cierra solo con el merge de su PR (`Closes #n`), nunca a mano | Todos | Sprint 2 |
| Toda tarea tiene su issue antes de empezarla | Todos | Sprint 2 |

## Seguimiento de las acciones del sprint anterior

| Acción del Sprint 0 | ¿Se cumplió? |
|---|---|
| Mergear TEC-01 y armar el CI (TEC-02) antes de empezar las historias | Sí: los dos el lunes 05/10 (#70 y #75) |
| Todo issue del sprint lleva SP antes de la Planning | Sí: todas las historias del sprint tenían SP al cerrar la Planning |
| Un PR no se aprueba sin el preview de Vercel en verde | Sí: los PRs mergeados tenían el preview en verde; en HU-11 (#88) la revisión documentó la prueba caso por caso |
| Daily por mensaje en el issue del sprint | Parcial: Carolina el día que correspondía, Mariano con atraso y Juan Pablo todas juntas al final (#86) |
