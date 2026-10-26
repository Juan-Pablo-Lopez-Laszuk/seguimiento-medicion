# Dailies · Sprint 3

Las dailies se escriben como comentario en el issue "Sprint 3 · Dailies" (#115): qué hice, qué voy a hacer, qué me
bloquea. Al cerrar el sprint se copian acá, una tabla por día. La columna "Registrada" indica cuándo se escribió
cuando no fue el mismo día.

## Lunes 19/10

| Integrante | Qué hice | Qué voy a hacer | Bloqueos | Registrada |
|---|---|---|---|---|
| Carolina | terminé HU-24 y HU-27 (las dos mergeadas) e hice HU-28, estado de un defecto (#112, espera la revisión de Mariano). | HU-33 (horas estimadas, reales y desviación), del camino crítico. | nada. | |
| Mariano | el domingo 18/10 revisé commit por commit, probé y mergeé HU-24 (#110) y HU-27 (#111): el Sprint 2 cerró en 33 SP. Hoy revisé HU-28 (#112): los RED fallan, los GREEN pasan y `defect` está al 100 %; el CI falla solo por un `errcheck` y dejé el arreglo en el PR. Leí la Review y la Retro del Sprint 2 (#113) y la Planning del Sprint 3 (#114). | mergear #112 con el CI en verde; arrancar HU-16b (#106, con la migración 0003) y TEC-06 (#69), después HU-09 y HU-10; confirmar las decisiones 3 y 4 de la Planning; corregir HU-06 en el modo memoria (borra una historia con horas o defectos, en PostgreSQL da 409). | nada; #112 espera el arreglo de Carolina. | |

## Martes 20/10

| Integrante | Qué hice | Qué voy a hacer | Bloqueos | Registrada |
|---|---|---|---|---|
| Carolina | se mergearon HU-28 (#112) y HU-33 (#116). Abrí HU-34 (#119) y HU-35 (#120), que esperan la revisión de Mariano. Revisé y mergeé HU-14 (#117) y #118 de Juan Pablo. | HU-29 (consultar defectos, #7) tarea por tarea, y atender #119 y #120. | nada. | |
| Juan Pablo | Carolina mergeó HU-14 (#117) y el plan del Sprint 4 (#118). Dividí las 48 HU y TEC en tareas (sub-issues `tipo:tarea` con horas), como pidieron los profesores. Terminé HU-15: #413 espera la revisión de Carolina. | HU-37 (#48) con sus tareas y después adelantar HU-38 (#49); cerrar las tareas de HU-34 y HU-35 cuando se mergeen. | nada. Le pregunté a Mariano por HU-16b, de la que depende Planning Poker. | |
| Mariano | HU-16b lista y en revisión (#414, con la migración 0003: hay que aplicarla en producción antes del merge). Mergeé HU-33 (#116). Revisé, probé y aprobé HU-34 (#119), HU-35 (#120) y HU-29 (#412); mergeé #119 y #120 y reabrí #412, que GitHub cerró al borrarse la rama de HU-34. Abrí TEC-06 parte 1 (#415): el CI falla si la cobertura de dominio y métricas baja del 80 % (hoy 99,5 %). | mergear #412; atender #414 y #415; la spec de TEC-05 (necesita la configuración de Supabase Auth); después HU-09, HU-10 y la matriz de TEC-06. | nada propio; #412 espera que Carolina resuelva un conflicto en `docs/ia/registro.md`. | |

## Miércoles 21/10

| Integrante | Qué hice | Qué voy a hacer | Bloqueos | Registrada |
|---|---|---|---|---|
| Carolina | resolví los conflictos de #119, #120 y #412 con `main`; Mariano mergeó HU-34 y HU-35. Aprobé HU-37 (#416). Terminé HU-26 (#4) y abrí #417 con sus tareas cerradas. A la noche: todas mis historias del Sprint 3 mergeadas (HU-28, HU-33, HU-34, HU-35, HU-29 y HU-26); mergeé HU-14, #118, HU-15 y HU-37 de Juan Pablo. HU-36 pasa al Sprint 4. | atender #412 y #417. | nada. | |
| Mariano | revisé, probé (también con PostgreSQL), aprobé y mergeé HU-29 (#412) y HU-26 (#417), con observaciones que no bloquean. Juan Pablo mergeó HU-16b (#414, migración 0003 en producción) y TEC-06 parte 1 (#415). TEC-05 lista y en revisión (#418): ocho ciclos RED → GREEN, token validado con el JWKS, vínculo por email; probada contra el Supabase de desarrollo y con el CI en verde. Cerré las tareas de TEC-05, HU-16b y TEC-06 que tenía abiertas. | atender #418; la matriz de TEC-06 y después HU-09 y HU-10; el arreglo de HU-06 en memoria. Lo que no llegue el viernes pasa al Sprint 4 (primero HU-10). | nada; #418 espera la revisión de Juan Pablo, y el día del merge hay que cargar las variables de Supabase en Production. | 22/10 |

## Jueves 22/10

| Integrante | Qué hice | Qué voy a hacer | Bloqueos | Registrada |
|---|---|---|---|---|
| Carolina | Mariano mergeó HU-29 (#412) y HU-26 (#417): terminé todas mis historias del Sprint 3. | revisar lo que suba Juan Pablo y atender lo que salga en la Review. | nada. | |
| Mariano | abrí la matriz de trazabilidad de TEC-06 (#419, 59 requerimientos con su test), el arreglo de HU-06 en memoria (#420) y HU-09 (#421, Subir y Bajar dentro de cada prioridad, 14 escenarios). Los tres con el CI en verde. | HU-10 (#23) en revisión el viernes; traer `main` a los PRs abiertos después de cada merge. | nada; Juan Pablo tiene para revisar #418, #419, #420 y #421 (primero #418). | |

## Viernes 23/10 y sábado 24/10

| Integrante | Qué hice | Qué voy a hacer | Bloqueos | Registrada |
|---|---|---|---|---|
| Juan Pablo | (del 21 al 24/10) mergeé HU-16b (#414) y TEC-06 parte 1 (#415); Carolina mergeó HU-37 (#416). Revisé commit por commit y probé TEC-05 (#418), la matriz de TEC-06 (#419), el arreglo de HU-06 (#420), HU-09 (#421) y HU-10 (#422): todos mergeados. Configuré Supabase y Vercel: el login anda en producción. Cargué a Carolina y Mariano en producción. Terminé HU-38 (#423, espera la revisión de Carolina). Subí la Planning del Sprint 4 (#425) y abrí las dailies del Sprint 4 (#424). | adelantar HU-39 (spike de la librería y spec). Planning Poker repartido: HU-17 y HU-18 Mariano, HU-19 Carolina y HU-22 Juan Pablo. | nada. | 24/10 |

Carolina y Mariano no dejaron daily el viernes 23 ni el sábado 24. El domingo 25/10 Carolina mergeó HU-38 (#423).
