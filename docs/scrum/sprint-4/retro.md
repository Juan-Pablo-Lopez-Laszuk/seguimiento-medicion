# Retrospectiva · Sprint 4 · 29/10/2026

Participantes: López, Piastrellini, Bravo · Facilitador: Juan Pablo

> Armada el 29/10 con el tablero, los PRs y las dailies del #424; se discute en la Retro. Lo marcado [A CONFIRMAR] se
> cierra ahí. Es la retro del último sprint: las acciones son para el cierre de la entrega y para el próximo proyecto.

## Métricas del sprint

- **SP comprometidos:** 23.
- **SP completados:** los 23 comprometidos y 2 adelantados (TEC-07b).
- **Velocidad:** 25 SP, con 4 días de código. Sprint 1: 39 · Sprint 2: 33 · Sprint 3: 48. **Total del proyecto: 145
  SP**, unos 36 por sprint.
- **Defectos abiertos:** 0.
- **No terminado:** HU-20 y HU-21 (6 SP, no comprometidas), que se dejaron afuera el 28/10.

| Integrante | Terminado | SP |
|---|---|---|
| Juan Pablo | HU-39, HU-22; seguridad en el pipeline (#431) | 7 de 7 |
| Mariano | modelo común de Planning Poker (#430), HU-17, HU-18 | 8 de 8 |
| Carolina | HU-36, HU-19 (con los arreglos de #438); adelantó TEC-07b | 8 de 8 (+2) |

## ¿Qué salió bien?

- **Planning Poker entró entero en tres días, aunque iba en cadena** (HU-17 → HU-18 → HU-19, con HU-22 al lado):
  - el modelo común (#430) se mergeó el lunes a la noche, con un archivo y una interfaz de repositorio por historia;
  - así cada uno pudo trabajar en paralelo sin pisarse;
  - Mariano dejó HU-17 y HU-18 en revisión un día antes de lo previsto.
- **Se completó todo lo comprometido** y lo que no iba a llegar (HU-20 y HU-21) se decidió el miércoles a la noche,
  antes del congelamiento. Fue la acción de la Planning: "si el miércoles algo viene atrasado, se decide qué queda
  afuera, empezando por HU-21 y HU-20".
- **Revisiones a fondo, que encontraron errores reales:** Mariano revisó HU-19 commit por commit y probándola. Encontró
  que el voto de alguien dado de baja aparecía al revelar y que las otras pestañas no se enteraban de la revelación.
  Carolina lo arregló ese mismo día (#438).
- **El 28/10 las revisiones fueron casi inmediatas:** HU-22 (#437), los arreglos de HU-19 (#438) y #439 se aprobaron en
  menos de una hora. 12 de los 15 PRs del sprint se aprobaron en menos de 24 horas.
- **Integrar antes de mergear:** HU-19 y HU-22 se probaron juntas en local antes de que entrara cualquiera de las dos.
  Apareció un nombre de error repetido (`ErrRondaSinRevelar`) y se resolvió antes de mergear, sin tocar el PR de
  Carolina.
- **Seguridad en el pipeline:**
  - `govulncheck` es check obligatorio de `main`;
  - Dependabot está activado;
  - se cerraron 15 vulnerabilidades de dependencias y la alerta de chi (#434).
- **Sin migraciones ni sorpresas en producción:** las tablas de Planning Poker estaban desde la 0001. El riesgo de la
  Planning ("una migración nueva") no hizo falta.
- **Las tareas se cerraron con el PR donde se hicieron** y el tablero quedó al día el mismo 28/10.
- **El Agile Enabler dejó la daily el mismo día el 28 y el 29/10**, y Mariano la dejó todos los días hasta el 28.

## ¿Qué podemos mejorar?

- **Tres revisiones pasaron las 24 horas:**
  - #431 (seguridad), unas 37: tenía conflicto con `main` en `docs/ia/registro.md`;
  - #426 (HU-36), unas 33;
  - #432 (HU-17), unas 27: la revisó Juan Pablo, otra vez.
- **Las dailies siguieron incompletas:**
  - Carolina dejó solo la del 28/10;
  - Juan Pablo cargó la del 26 y la del 27 juntas, el 27 a la noche.

  [A CONFIRMAR] las del jueves 29 de Mariano y Carolina.
- **Conflictos repetidos en los mismos archivos:**
  - `docs/ia/registro.md`, en casi todos los PRs (cada uno suma una fila arriba);
  - la página de Planning Poker (`poker.html`), `router.go` y `features_test.go`, al juntar HU-18, HU-19 y HU-22.

  Una fila de HU-19 quedó en el medio del registro y no arriba.
- **El trabajo se amontonó el último día de código:** el 28/10 se mergearon 7 PRs y el 29/10 entró el proyecto de
  ejemplo (#440), que es el plan B de la demo, sin margen para observaciones.
- **La documentación quedó para después del congelamiento:** las secciones 1 y 3 del manual, la cobertura y las
  evidencias de TDD del informe, la arquitectura y el guion. La Planning pedía documentar cada historia al terminarla.
- **Las horas reales de las tareas** [A CONFIRMAR] si se anotaron al cerrarlas, como pedía la acción del Sprint 3.
- **El PostgreSQL local de pruebas se corrompió** porque estaba en la carpeta temporal de Windows. Hubo que rehacerlo
  y se perdieron los datos de prueba.

## Acciones

Para el cierre de la entrega (30/10 al 02/11):

| Acción | Responsable | Para cuándo |
|---|---|---|
| La cobertura y las evidencias de TDD del informe | Mariano | Vie 30/10 |
| El PR de documentación (#442: el manual de las secciones 1 a 3, la arquitectura, el guion y la guía) y el de este cierre | Juan Pablo | Vie 30/10 |
| Revisar y mergear #440 (proyecto de ejemplo) | Carolina | Vie 30/10 |
| Los datos del Sprint 4 en el informe de métricas | Carolina [A CONFIRMAR] | Sáb 31/10 |
| Prueba completa en producción con las tres cuentas y el ensayo de la presentación, con el plan B listo | Los tres | Sáb 31/10 o dom 01/11 |
| Crear las cuentas de los profesores con "Add user" y sumarlos como integrantes de un proyecto | Juan Pablo | Antes del lunes 02/11 |

Para el próximo proyecto:

| Acción | Por qué |
|---|---|
| Un registro de IA por historia (un archivo cada uno) o filas al final, en lugar de una tabla con la fila arriba | Es el archivo que más conflictos dio en los cuatro sprints |
| Cuando dos historias tocan la misma pantalla, el segundo PR se prueba junto con el primero antes de pedir revisión | Así se encontraron a tiempo los conflictos de HU-19 y HU-22 |
| La documentación de cada historia va en su mismo PR | Para no dejarla para después del congelamiento |
| La base de pruebas local fuera de la carpeta temporal | Windows la limpia y la corrompe |

## Seguimiento de las acciones del sprint anterior

| Acción del Sprint 3 | ¿Se cumplió? |
|---|---|
| Con dos o más PRs esperando, Juan Pablo revisa antes de seguir con su código. Ninguna revisión pasa las 24 horas | Parcial: #430 y #433 en menos de 8 horas, pero #432 tardó unas 27 |
| Daily el mismo día, también la del Agile Enabler | Parcial: Juan Pablo el 28 y el 29; Carolina solo el 28 |
| Todo lo que entra al congelamiento está en revisión el miércoles 28/10 a la noche | Sí, salvo el proyecto de ejemplo (#440, plan B de la demo, sin SP), que entró el 29/10 |
| Antes de mergear la base de un PR apilado, se le cambia la base a `main` al que depende de ella | Sí: #433 estaba apilado sobre #432 y no se cerró |
| El autor del PR que se mergea primero avisa, y los demás traen `main` ese mismo día | Sí: HU-22 trajo `main` con HU-19 el mismo día del merge (#437) |
| Al cerrar cada tarea se anotan las horas que llevó | [A CONFIRMAR] |
| Un PR aprobado lo mergea el revisor en el momento | Sí: todos los PRs aprobados durante el sprint se mergearon el mismo día de la aprobación |
