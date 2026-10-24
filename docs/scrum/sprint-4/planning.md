# Sprint Planning · Sprint 4 · 26/10/2026

Participantes: López, Piastrellini, Bravo · Facilitador: Juan Pablo

> Borrador armado el 24/10 a partir del tablero, la [Planning del Sprint 3](../sprint-3/planning.md) y las horas de las
> tareas de cada historia. Lo marcado [A CONFIRMAR] lo cierra el equipo comentando en el PR de esta acta.

## Fechas

Del **lunes 26/10 al domingo 01/11** (7 días). Es el sprint de cierre: rendimos el **lunes 02/11**.

**El código se congela el jueves 29/10** (plan del Sprint 3). Del 26 al 29 entran funcionalidades. El viernes 30 y el
fin de semana son solo para bugs, documentación y el ensayo de la presentación.

## Sprint Goal

> Cerrar el producto para la entrega: Planning Poker de punta a punta (abrir sesión, votar oculto, revelar y registrar
> la estimación), los gráficos de las métricas y la exportación del reporte a PDF; y completar la documentación, el
> informe de métricas y la presentación final.

[A CONFIRMAR]

## Cómo terminó el Sprint 3

- **Comprometido: 40 SP, entregado: 40 SP.** Están en `main` las doce historias comprometidas.
- **Se adelantaron 5 SP:** HU-26 y HU-29 (Carolina). HU-38 (Juan Pablo, 3 SP) está en revisión (#423). Si se mergea
  antes del domingo, el Sprint 3 cierra con **48 SP**.
- La matriz de trazabilidad (TEC-06) queda en **51 de 59** requerimientos cubiertos. Los 8 que faltan son los de este
  sprint: Planning Poker (HU-17 a HU-22), gráficos (HU-36) y PDF (HU-39).

## Capacidad y velocidad

- Velocidad: **39 SP** (Sprint 1), **33 SP** (Sprint 2) y **45 a 48 SP** (Sprint 3). El promedio es de unos **40 SP**
  por semana.
- Este sprint tiene **4 días de código** (lunes a jueves), así que la capacidad para funcionalidades ronda los
  **22 SP**. El resto de la semana va a documentación y presentación.
- Se proponen **23 SP** comprometidos, que es lo imprescindible del plan del Sprint 3: Planning Poker sin HU-20 ni
  HU-21 (13 SP), HU-36 (5) y HU-39 (5).

### Reparto: Planning Poker no entra entero en una persona

Las tareas de las cuatro historias imprescindibles de Planning Poker suman **44,5 h**. Si las hace solo Mariano, son 11 h
por día de lunes a jueves, y además van en cadena: si una se atrasa, se atrasan todas. Por eso se propone repartirlas:

| Persona | Historias | Horas estimadas |
|---|---|---|
| Mariano | HU-17 (11,5 h) y HU-18 (16 h) | 27,5 h |
| Carolina | HU-36 (16,5 h) y HU-19 (9,5 h) | 26 h |
| Juan Pablo | HU-39 (15 h) y HU-22 (7,5 h) | 22,5 h, más la coordinación, las actas y el cierre de HU-38 |

Para poder trabajar en paralelo, Mariano escribe el lunes **un modelo común** para la épica: sesión, ronda y voto, con
los nombres de las tablas. Lo pone en la spec de HU-17 y los demás construyen sobre esa base. La escala de votación es
la de HU-16.

Reparto aprobado por Juan Pablo el 24/10, para no cargar a Mariano y trabajar en paralelo. [A CONFIRMAR] Mariano y Carolina.

## Historias del sprint

### Comprometidas

| Orden | Historia | SP | Horas | Responsable | Depende de | Por qué en este orden |
|---|---|---|---|---|---|---|
| 0 | HU-38 · Reporte de proyecto (#49) | 3 | — | Juan Pablo | HU-37 (hecha) | Solo si no se mergea antes del domingo. Ya está en revisión (#423); la revisa Carolina el lunes a primera hora, porque HU-39 la necesita |
| 1 | HU-17 · Abrir sesión de Planning Poker (#25) | 3 | 11,5 | Mariano | TEC-05, HU-16 (hechas) | Camino crítico de la épica: trae el modelo de sesión que usan las otras tres |
| 1 | HU-36 · Gráficos de métricas (#14) | 5 | 16,5 | Carolina | HU-35 (hecha) | No depende de nadie: arranca el lunes |
| 1 | HU-39 · Exportar reporte a PDF (#50) | 5 | 15 | Juan Pablo | HU-38 | El spike de la librería (T1, #375) se puede hacer el lunes mientras se revisa HU-38 |
| 2 | HU-18 · Votar de forma oculta (#26) | 5 | 16 | Mariano | HU-17 | Es la más grande: conviene que su spec esté revisada el martes |
| 2 | HU-22 · Registrar la estimación acordada (#30) | 2 | 7,5 | Juan Pablo | HU-17, HU-16 (hecha) | Solo necesita la sesión: va en paralelo con HU-18 |
| 3 | HU-19 · Revelar estimaciones (#27) | 3 | 9,5 | Carolina | HU-18 | Necesita los votos. Se puede escribir sobre el modelo de la spec de HU-18 y mergear cuando HU-18 esté en `main` |

**Total comprometido: 23 SP** (26 si HU-38 pasa al Sprint 4) · Mariano 8 · Carolina 8 · Juan Pablo 7.

### Si sobra tiempo (no comprometidas)

| Historia | SP | Horas | Condición |
|---|---|---|---|
| HU-20 · Detectar diferencias entre estimaciones (#28) | 3 | 8 | Después de HU-19. Es lógica de dominio pura: la toma el primero que termine |
| HU-21 · Realizar nuevas rondas (#29) | 3 | 9,5 | Después de HU-19 |
| TEC-07b · Componentes de formularios y tablas (#101) | 2 | 6 | Carolina, después de HU-36 y HU-19 |

Lo que no esté en revisión el **jueves 29/10 a la noche** no entra en la entrega: queda en el Product Backlog y se
cuenta en la Review como trabajo pendiente.

## Documentación de la entrega

La consigna pide 13 entregables. Del 1 al 9 (repositorio, tablero, backlogs, specs, escenarios, pruebas, código,
evidencias de TDD y software funcionando) ya están en el repositorio y en el tablero. Faltan estos:

| Entregable | Qué hay que hacer | Responsable | Para cuándo |
|---|---|---|---|
| 10 · Informe de métricas y cobertura | `docs/informe-metricas.md`. Métricas del propio equipo por sprint: SP planificados y completados, velocidad, horas estimadas y reales, defectos. Salen del tablero y de las tareas | Carolina | Vie 30/10 |
| 10 · Informe de métricas y cobertura | La sección de cobertura: cobertura por paquete, el umbral de TEC-06 y cómo se mide | Mariano | Vie 30/10 |
| 8 · Evidencias de TDD | Una sección del informe que explique cómo ver los commits `RED` → `GREEN` de cada historia, con dos o tres ejemplos | Mariano | Vie 30/10 |
| 11 · Actas de retrospectivas | Retro del Sprint 3 (después de la Review) y del Sprint 4 (domingo 01/11) | Juan Pablo | Dom 01/11 |
| 12 · Documentación técnica | `docs/arquitectura.md`, ya en borrador, más revisar `docs/deploy.md` con lo de TEC-05 | Juan Pablo | Vie 30/10 |
| 12 · Manual breve de usuario | `docs/manual-usuario.md`, una sección por épica. Juan Pablo: proyectos, sprints y reportes. Mariano: login, backlog y estimación (Planning Poker). Carolina: esfuerzo, defectos, métricas y panel | Los tres | Vie 30/10 |
| 13 · Presentación y demostración | Guion, diapositivas y un proyecto de ejemplo cargado en producción para la demo | Los tres; coordina Juan Pablo | Ensayo el sáb 31/10 o el dom 01/11 |

[A CONFIRMAR] La división de la documentación (venía propuesta del Sprint 3).

## Calendario

| Día | Qué pasa |
|---|---|
| Lun 26/10 | Planning. Carolina revisa HU-38. Cada uno escribe la spec y los escenarios de su primera historia (Mariano también el modelo común de Planning Poker) y los revisamos en el día |
| Mar 27/10 | HU-17 en revisión. Mariano empieza HU-18 y Juan Pablo empieza HU-22 sobre HU-17. HU-36: endpoint JSON con sus tests |
| Mié 28/10 | HU-18 y HU-39 en revisión. HU-36: los gráficos. Carolina escribe HU-19 sobre el modelo de HU-18 |
| Jue 29/10 | **Congelamiento.** HU-19, HU-22 y HU-36 en revisión antes de la noche. Se revisa y se mergea en el día |
| Vie 30/10 | Solo bugs. Se cierra la documentación y el informe de métricas. Prueba completa en producción |
| Sáb 31/10 · Dom 01/11 | Ensayo de la presentación, Review y Retro del Sprint 4, actas |
| Lun 02/11 | Presentación y demostración final |

## Decisiones

| # | Decisión | Propuesta | ¿Confirmada? |
|---|---|---|---|
| 1 | Planning Poker: quién vota | Vota el integrante vinculado a la cuenta con la que se entró (TEC-05). Sin login (modo memoria y preview) se elige el integrante en la pantalla. No hay tiempo real: la página se actualiza al recargar | [A CONFIRMAR] Mariano, en la spec de HU-17 |
| 2 | HU-39: cómo se genera el PDF | En el servidor, con una librería de Go sin dependencias externas, porque CA-39.1 pide descargar un archivo con nombre y CA-39.3 pide menos de 10 s en Vercel. Lo confirma el spike (#375) | [A CONFIRMAR] Juan Pablo, en el spike |
| 3 | HU-39: qué se exporta | El reporte de un sprint (`proyecto-sprint-fecha.pdf`, CA-39.1) y, si entra, el del proyecto. El contenido es el de HU-37 y HU-38 sin cambios (CA-39.2) | [A CONFIRMAR] Juan Pablo, en la spec |
| 4 | HU-36: librería de gráficos | Una librería de JavaScript cargada por CDN, sin paso de build, con los datos de un endpoint JSON con tests (CA-36.2). El burndown usa la fecha en que se completó cada historia (HU-13) | [A CONFIRMAR] Carolina, en la spec |
| 5 | Quién puede entrar a la app | Una vez registrados el equipo y los profesores, se cierra el registro de cuentas nuevas en Supabase. En la Review se presenta como limitación conocida que cualquier cuenta registrada ve todos los proyectos (lo planteó Mariano) | [A CONFIRMAR] Juan Pablo |
| 6 | Congelamiento | Después del jueves 29/10 solo entran arreglos de bugs y documentación | Viene del plan del Sprint 3 |

## Antes de arrancar (pendientes del Sprint 3)

- **Cerrar las tareas de historias ya mergeadas:** #255 (HU-09), #260 (HU-10) y #369 (HU-37) se cerraron el 24/10.
  Cuando se mergee HU-38 se cierran #370 a #374.
- **Tablero:** pasar HU-17 a HU-22 y HU-36 de "Sprint 3" a "Sprint 4 · Cierre", y ponerle sprint y SP a TEC-07b.
  Cambiar el responsable de HU-19 (Carolina), HU-22 (Juan Pablo) y sus tareas (#283 a #289 y #303 a #309).
- **Dailies:** copiar las del issue #115 a [`../sprint-3/dailies.md`](../sprint-3/dailies.md) y cerrar #115. El de
  este sprint ya está abierto: #424.
- **Review y Retro del Sprint 3:** [A CONFIRMAR] la fecha con el profesor.
- **Producción:** Carolina y Mariano ya están cargados como integrantes (24/10), así pueden votar en Planning Poker.

## Acciones que siguen

- **Revisiones en menos de 24 horas**, y esta semana en el día: Juan Pablo revisa a Mariano, Mariano a Carolina y
  Carolina a Juan Pablo. Si el revisor no puede, avisa en el issue de dailies y revisa el siguiente de la rotación.
- **Después de cada merge, los PRs abiertos traen `main` en el día** (lo hace el autor). Esta semana importa más:
  HU-17, HU-18, HU-19 y HU-22 tocan el mismo paquete.
- **Cada historia actualiza su fila en la matriz de trazabilidad** (`docs/trazabilidad-requisitos.md`) a "Cubierto" en
  su mismo PR, con sus escenarios y tests. El test de TEC-06 controla que lo que nombra la fila exista.
- **Cada uno documenta lo suyo al terminarlo**: su sección del manual en el mismo PR de la historia, o en uno aparte
  el viernes.

## Riesgos y dependencias

- **Planning Poker va en cadena** (HU-17 → HU-18 → HU-19). Si HU-17 no está en `main` el martes, HU-19 no llega al
  jueves. Para que no pase: el modelo común se escribe el lunes, HU-17 se revisa apenas entra y cada PR es chico.
- **Una migración nueva:** las tablas de sesiones, rondas y votos van en una migración que Juan Pablo aplica en
  producción el mismo día del merge (como la 0003 de HU-16b). Conviene que sea una sola migración para toda la épica
  y que entre con HU-17.
- **HU-39 depende de HU-38**, y del tiempo de respuesta de Vercel (10 s). El spike del lunes lo resuelve antes de
  escribir código.
- **HU-36 trae JavaScript por primera vez** (hasta ahora la app es solo HTML del servidor). Tiene que andar en el
  celular y no romper la impresión del reporte.
- **No hay margen:** después del 29/10 no hay otro sprint. Si el miércoles a la noche algo está atrasado, se decide en
  la daily del jueves qué se deja afuera, empezando por HU-21 y HU-20.

## Dailies

Lunes a viernes, como comentario en el issue "Sprint 4 · Dailies" (#424). Al cerrar el sprint se copian a
[`dailies.md`](dailies.md).
