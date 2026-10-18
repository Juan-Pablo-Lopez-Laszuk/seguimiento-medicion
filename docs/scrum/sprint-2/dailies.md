# Dailies · Sprint 2

Las dailies se escriben como comentario en el issue "Sprint 2 · Dailies" (#95): qué hice, qué voy a hacer, qué me
bloquea. Al cerrar el sprint se copian acá, una tabla por día. La columna "Registrada" indica cuándo se escribió
cuando no fue el mismo día.

## Martes 13/10

| Integrante | Qué hice | Qué voy a hacer | Bloqueos | Registrada |
|---|---|---|---|---|
| Carolina | revisé, probé en el preview, aprobé y mergeé HU-04 (#97) y HU-02 (#98). Aprobé la Review y Retro del Sprint 1 (#94, con dos correcciones chicas) y la Planning del Sprint 2 (#96). TEC-04 cerrada: `/health` de producción ya dice `"base":"ok"`. | TEC-07 (diseño base) en un PR chico hoy, para que `base.html` llegue temprano. | nada. | |
| Juan Pablo | HU-04 (#97) y HU-02 (#98) mergeadas, revisadas por Carolina. Dejé revisadas las specs y los escenarios de HU-12 y HU-13. Corregí lo que pidió Carolina en la Review y la Retro (#94). | HU-12 apenas estén en `main` HU-07, HU-16 y HU-08; mientras tanto, coordinar con Mariano cómo se comparten los estados de las historias. | HU-12 y HU-13 dependen de HU-07, HU-16 y HU-08 (Mariano). | |
| Mariano | aprobé y mergeé TEC-07 (#99) después de probarlo con ancho de celular (las observaciones quedaron en #101). Confirmé en el #95 las decisiones 5, 6 y 8 de la Planning. Hice HU-07 (#20) y abrí el PR #100. Le consulté a Juan Pablo por el historial de HU-16 (CA-16.3). | HU-08 y HU-16, el camino crítico para HU-12 y HU-13. | #100 espera la revisión de Juan Pablo; HU-16 espera su decisión sobre el historial. | 14/10 |

## Miércoles 14/10

| Integrante | Qué hice | Qué voy a hacer | Bloqueos | Registrada |
|---|---|---|---|---|
| Carolina | TEC-07 mergeada (#99, revisada por Mariano). Abrí #101 con lo que falta de formularios y tablas. Mergeé la Review y Retro del Sprint 1 (#94) y la Planning (#96) después de las correcciones. | HU-25 (#3), que después usa HU-24. | nada. | |
| Mariano | HU-07 (#100) quedó mergeada. Revisé, aprobé y mergeé HU-25 (#102). Hice HU-08 (#21) y abrí el PR #103. | HU-16 apenas Juan Pablo decida lo del historial, y después HU-06. | #103 espera la revisión de Juan Pablo y de eso dependen HU-12 y HU-13; HU-16 espera la decisión sobre el historial. | |
| Juan Pablo | aprobé y mergeé HU-07 (#100) la noche anterior. Ajusté las specs de HU-12 y HU-13 para usar lo que trajo HU-07. | revisar HU-08 (#103) apenas la abra Mariano y arrancar HU-12. | HU-12 y HU-13 esperan HU-08 (Mariano). | 17/10 |

## Jueves 15/10

| Integrante | Qué hice | Qué voy a hacer | Bloqueos | Registrada |
|---|---|---|---|---|
| Mariano | HU-08 (#103) quedó mergeada. Con la decisión de Juan Pablo dividí HU-16: creé HU-16b (#106) y dejé #24 con CA-16.1 y CA-16.2. Hice HU-16 y abrí el PR #108 pasada la medianoche. | HU-06 (#19) y, si sobra tiempo, HU-16b (#106). | #108 espera la revisión de Juan Pablo. HU-16b necesita la migración 0003. | 16/10 |
| Juan Pablo | aprobé y mergeé HU-08 (#103). Decidí dividir HU-16 (el historial pasa a HU-16b). Agregué el workflow que pide la revisión según la rotación (#104). Terminé HU-12 (#105) y HU-13 (#107, apilado sobre #105). | atender la revisión de Carolina en #105 y #107, y revisar HU-16. | nada; #105 y #107 esperan la revisión de Carolina. | 17/10 |
| Carolina | (registrada junto con el 16 y el 17/10) HU-25 mergeada (#102). Por error mergeé el #104 antes de terminar de revisarlo. | revisar HU-12 y HU-13. | no pude trabajar por un problema con mi herramienta. | 17/10 |

## Viernes 16/10

| Integrante | Qué hice | Qué voy a hacer | Bloqueos | Registrada |
|---|---|---|---|---|
| Mariano | revisé el repo: no tenía PRs de Carolina para revisar. Hice HU-06 (#19) y abrí el PR #109. | atender la revisión de #109; si sobra tiempo, HU-16b (#106). | #108 y #109 esperan la revisión de Juan Pablo. | |

Sin daily el mismo día: Juan Pablo y Carolina (las cargaron el 17/10, juntas con la del sábado).

## Sábado 17/10

| Integrante | Qué hice | Qué voy a hacer | Bloqueos | Registrada |
|---|---|---|---|---|
| Carolina | revisé, aprobé y mergeé HU-12 (#105) y HU-13 (#107), probados commit por commit (171 escenarios en verde). | HU-24 (#2) y HU-27 (#5), lo que me queda del sprint. | nada; voy atrasada, lo que no llegue lo aviso en la Review. | (la tituló 18/10) |
| Juan Pablo | (junto con el 16/10) Carolina mergeó HU-12 y HU-13: mi parte del sprint está completa (12 SP). Revisé HU-16 (#108) y HU-06 (#109): resolví los conflictos con HU-12 y HU-13, las probé integradas (221 escenarios en verde, también contra PostgreSQL) y las mergeé. | preparar la Review y la Retro y cerrar el sprint. | nada; HU-24 (#110) y HU-27 (#111) esperan la revisión de Mariano. | |

Sin daily: Mariano.
