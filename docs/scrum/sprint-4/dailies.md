# Dailies · Sprint 4

Las dailies se escriben como comentario en el issue "Sprint 4 · Dailies" (#424): qué hice, qué voy a hacer, qué me
bloquea. Al cerrar el sprint se copian acá, una tabla por día. La columna "Registrada" indica cuándo se escribió
cuando no fue el mismo día.

## Lunes 26/10

| Integrante | Qué hice | Qué voy a hacer | Bloqueos | Registrada |
|---|---|---|---|---|
| Juan Pablo | se mergearon HU-39 (#428, exportar a PDF), la Planning del Sprint 4 (#425) y la Review, la Retro y las dailies del Sprint 3 (#429). Revisé el modelo común de Planning Poker (#430). Abrí #431 (seguridad): `govulncheck` en el CI, pgx v5.11.0, x/text v0.43.0 y Go 1.27.3, que cierran 15 vulnerabilidades. Activé Dependabot. | HU-22 sobre HU-17. | nada. | 27/10 (junto con la del 27) |
| Mariano | aprobé HU-36 (#426) y el informe de métricas y el manual (#427). Confirmé el reparto y el modelo común en #425: quedaron mergeados los tres. Subí el modelo común de Planning Poker (#430: spec de HU-17, feature y tipos `Sesion`, `Ronda` y `Voto`; sin migración, las tablas están desde la 0001). Subí HU-17 (#432, abrir una sesión), un día antes de lo previsto. | martes: la spec y la feature de HU-18 a primera hora y después su TDD. Miércoles: HU-18 en revisión y revisar HU-19. | nada; Juan Pablo tiene #430 y #432 para revisar. | |
| Carolina | — | — | — | no dejó daily |

## Martes 27/10

| Integrante | Qué hice | Qué voy a hacer | Bloqueos | Registrada |
|---|---|---|---|---|
| Juan Pablo | revisé, probé en local y mergeé HU-17 (#432) y HU-18 (#433). Resolví el conflicto de #431 con `main`. Terminé HU-22 en local: spec, 10 escenarios y 5 ciclos RED → GREEN. | rebasar HU-22 sobre `main` y dejarla en revisión (la revisa Carolina); sumar govulncheck como check obligatorio cuando se mergee #431. | nada para HU-22; para usarla en la app hace falta HU-19 (revelar). Le pregunté a Carolina cómo venía HU-19. | 27/10 a la noche |
| Mariano | respondí la revisión de #432 con un escenario nuevo (CA-17.3). Subí HU-18 (#433, votar de forma oculta): se ve quién votó y nunca los valores, la lista se actualiza sola cada 5 segundos con HTMX y el repositorio solo devuelve quiénes votaron. 7 ciclos RED → GREEN, un día antes de lo previsto. | atender lo que se pida en #432 y #433, revisar HU-19 y cerrar las tareas de HU-17 y HU-18 al mergear. | nada; Juan Pablo tiene #432 y #433, y Carolina tiene en #433 la sección "Para HU-19". | |
| Carolina | — | — | — | no dejó daily |

## Miércoles 28/10

| Integrante | Qué hice | Qué voy a hacer | Bloqueos | Registrada |
|---|---|---|---|---|
| Juan Pablo | se mergearon #431 (seguridad) y HU-22 (#437). Aprobé y mergeé #434 (Dependabot, chi 5.4.0), que cerró la única alerta. Resolví los conflictos de HU-22 con HU-19 después de probarlas juntas en local. Cerré las tareas de HU-17, HU-18, HU-22 y TEC-07b, y dejé el tablero al día. La matriz de trazabilidad quedó en 57 de 59. | sumar govulncheck como check obligatorio; mi sección del manual. | nada. Propuse dejar afuera HU-20 y HU-21. | |
| Mariano | se mergearon HU-17 (#432) y HU-18 (#433). Revisé HU-19 (#435) commit por commit: encontré que el voto de alguien dado de baja aparecía al revelar y que las otras pestañas no se enteraban de la revelación. Aprobé los arreglos (#438) y TEC-07b (#436), probados en la app y en el celular. | jueves: cierre del sprint. Viernes: cobertura y evidencias de TDD del informe, y sus secciones del manual. | nada; #438 aprobado para mergear. | |
| Carolina | aprobé y mergeé #431. Subí HU-19 (#435, mergeada por Mariano) y cerré sus tareas. Revisé, aprobé y mergeé HU-22 (#437). Hice TEC-07b (#436) y los arreglos que pidió Mariano para HU-19 (#438). | que se mergeen #436 y #438: con eso termina lo suyo del sprint. | nada. | |

A la noche, Juan Pablo y Mariano decidieron dejar afuera del sprint HU-20 y HU-21: no estaban comprometidas y nadie las
había empezado (comentado en #28 y #29).

## Jueves 29/10 · congelamiento del código

| Integrante | Qué hice | Qué voy a hacer | Bloqueos | Registrada |
|---|---|---|---|---|
| Juan Pablo | HU-20 y HU-21 volvieron al Product Backlog sin sprint. "Vulnerabilidades (govulncheck)" es check obligatorio de `main`. Creé la cuenta de Carolina en Supabase y probé Planning Poker en producción con login. Abrí #440 (proyecto de ejemplo con `go run ./cmd/ejemplo`, plan B de la demo), probado en local. Mi sección del manual y el borrador de la Review y la Retro, listos. | viernes: PR de documentación (manual, arquitectura, guion) y el de la Review, la Retro y las dailies. Fin de semana: ensayo. | nada; Carolina tiene #440 para revisar. | |
| Mariano | [A COMPLETAR] | | | |
| Carolina | [A COMPLETAR] | | | |
