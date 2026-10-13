# Dailies · Sprint 1

Las dailies se escriben como comentario en el issue "Sprint 1 · Dailies" (#86): qué hice, qué voy a hacer, qué me
bloquea. Al cerrar el sprint se copian acá, una tabla por día. La columna "Registrada" indica cuándo se escribió
cuando no fue el mismo día.

## Lunes 05/10

| Integrante | Qué hice | Qué voy a hacer | Bloqueos | Registrada |
|---|---|---|---|---|
| Juan Pablo | mergeé TEC-01 (#70) después de aprobar el arreglo de Carolina para Vercel (#73), y armé e integré el CI (TEC-02, #75). Hice HU-01 completa con TDD (spec, escenarios, dominio, repositorio en memoria, caso de uso y pantallas) y los borradores de la Review y la Retro del Sprint 0 (#74). | cerrar la spec de HU-01 y abrir su PR; ajustar las actas con la revisión de Mariano. | nada. | 12/10 |
| Mariano | revisé, aprobé y mergeé el PR #72 (pasos de godog de HU-31) de Carolina. | revisar HU-31 (#76) y cerrar los story points que faltaban en las tareas del Sprint 0. | nada. | 09/10 |

Sin daily: Carolina.

## Martes 06/10

| Integrante | Qué hice | Qué voy a hacer | Bloqueos | Registrada |
|---|---|---|---|---|
| Juan Pablo | dejé la spec de HU-01 en estado Implementada con la trazabilidad y su registro de IA. Cargué los story points de mis tareas del Sprint 0 y corregí las actas con lo que pidió Mariano en el #74 (escala sin "?"). | abrir el PR de HU-01 y armar la Planning del Sprint 1. | nada. | 12/10 |
| Mariano | aprobé y mergeé HU-31 (#76). Cargué a posteriori los story points de #64, #65 y #66 y los reflejé en la Review y la Retro del Sprint 0 (PR #77), y dejé anotado que el 80 % de cobertura es una meta interna. Revisé las actas del Sprint 0 de Juanpa (#74). | terminar de revisar #74 y HU-30 (#79). | TEC-03 y mis historias esperan la Planning del Sprint 1, que arma Juanpa. | 09/10 |

Sin daily: Carolina.

## Miércoles 07/10

| Integrante | Qué hice | Qué voy a hacer | Bloqueos | Registrada |
|---|---|---|---|---|
| Juan Pablo | se mergeó HU-01 (#78) y las actas del Sprint 0 (#74). Revisé y aprobé #81 y #82 de Mariano. Armé la Planning del Sprint 1 con el recorte por dependencias y pasé `docs/metricas.md` a v1.0. | subir la Planning y arrancar HU-11. | nada. | 12/10 |
| Mariano | aprobé y mergeé HU-30 (#79) y las actas del Sprint 0 (#74). Armé el registro de uso de IA de Mariano (#81) y dejé el modelo de datos como "acordado" (#82), con sus issues #83 y #84. | TEC-03 (base de datos y migraciones) y revisar lo que suba Carolina. | la conformidad de Carolina sobre Tarea, RegistroEsfuerzo y Defecto sigue pendiente. | 09/10 |

Sin daily: Carolina.

## Jueves 08/10

| Integrante | Qué hice | Qué voy a hacer | Bloqueos | Registrada |
|---|---|---|---|---|
| Juan Pablo | se mergeó la Planning (#85); creé este issue de dailies (#86) y dejé el tablero igual que el acta. Hice HU-11 (crear sprint) y HU-03 (registrar integrantes) con TDD, con las specs revisadas antes del código. | abrir los PRs de HU-11 y HU-03 (Carolina revisa). | nada. | 12/10 |
| Carolina | HU-30 (story points) y HU-31 (velocidad) mergeadas; revisé HU-01 y TEC-02 de Juanpa y la Planning (#85). | hoy HU-32 (porcentaje de completadas); después HU-23 y TEC-04. | nada por ahora. TEC-04 depende de TEC-03 (base de datos). | el mismo día |

Sin daily: Mariano.

## Viernes 09/10

| Integrante | Qué hice | Qué voy a hacer | Bloqueos | Registrada |
|---|---|---|---|---|
| Juan Pablo | se mergearon HU-11 (#88) y HU-03 (#90); resolví el conflicto del #90 con #89. Revisé TEC-03 (#91) contra una base real y pedí dos arreglos (Vercel no subía `migrations/` y el linter). Escribí las specs en borrador de HU-04 y HU-02 y el refinamiento del Sprint 2. | aprobar TEC-03 cuando esté en verde y publicar la decisión del #80. | nada. El riesgo del sprint es HU-05, que no empezó. | 12/10 |
| Mariano | revisé, aprobé y mergeé HU-23 (#89) y HU-32 (#87). Hice TEC-03: migración de las 12 tablas, repositorios de proyecto, sprint e integrante, y comando `migrar`; la migración está aplicada en Supabase dev. Abrí el PR #91 y respondí la revisión de Juanpa (Vercel y linter). | sábado 10/10: atender lo que quede de #91 y terminar lo que falta del sprint (HU-05 y #80). | #91 espera la aprobación de Juanpa y TEC-04 de Carolina depende de que se mergee. La #80 necesita decisiones de Juanpa. HU-05 todavía no tiene spec y el sprint cierra el domingo. | el mismo día |
| Carolina | terminé HU-23 (horas estimadas, la regla) y abrí el PR #89. Revisé, probé en el preview y aprobé HU-11 de Juanpa (#88). | esperar la revisión de #87 y #89, y seguir con TEC-04 cuando esté TEC-03. | TEC-04 depende de TEC-03 (base de datos). El #87 y el #89 esperan la revisión de Mariano. | el mismo día |

## Sábado 10/10

| Integrante | Qué hice | Qué voy a hacer | Bloqueos | Registrada |
|---|---|---|---|---|
| Juan Pablo | aprobé y mergeé TEC-03 (#91). Publiqué la decisión del #80. Dejé los borradores de la Review y la Retro del Sprint 1. | revisar las specs de HU-04 y HU-02 para tenerlas listas para la Planning, y cerrar el sprint el domingo (Review y Retro). | nada. | 12/10 |
| Mariano | Confirmé con Juan Pablo la decisión de #80 y la cerré. Dejé #91 (TEC-03) con el CI en verde y Juan Pablo lo aprobó y mergeó. Hice HU-05 (crear ítem del Product Backlog, #18): spec, escenarios BDD, dominio, repositorios en memoria y PostgreSQL, caso de uso y pantallas, con ciclos RED → GREEN, y abrí el PR #93 (CI en verde). Revisé, aprobé y mergeé el PR #92 de Carolina (TEC-04, `/health` con chequeo de la base), con `Refs #15`. | Domingo 11/10: atender la revisión de Juan Pablo sobre #93 y llegar a la Review y la Retro del Sprint 1. Después, la Planning del Sprint 2, donde entrarían HU-06, 07, 08 y 16. | #93 espera la revisión de Juan Pablo. #15 (TEC-04) no se puede cerrar hasta que Juan Pablo aplique las migraciones en producción y cargue `DATABASE_URL` en Vercel. | el mismo día |

## Domingo 11/10 y lunes 12/10

No hay dailies. El lunes 12/10 fue el día extra del sprint para terminar HU-05 y TEC-04 (ver la
[Review](review.md)).
