# Manual de usuario

Entregable 12 de la consigna. Una sección por épica, con los pasos para usar cada pantalla de la aplicación.

> Secciones 1 a 3: Juan Pablo. Secciones 4 a 6: Carolina.

## 1. Ingresar a la aplicación

La aplicación está en https://seguimiento-medicion.vercel.app. Todas las páginas piden haber iniciado sesión. Si
entrás a una sin sesión, aparece **Iniciar sesión** con el aviso "Iniciá sesión para continuar." y, después de
ingresar, la app te lleva a la página que querías ver.

### 1.1 La cuenta

El registro está cerrado: las cuentas las crea el Agile Enabler desde Supabase, con el **mismo email** con el que te
registró como integrante del proyecto (sección 2.3). Ese email es el que vincula tu cuenta con tu integrante, y así la
app sabe quién sos en cada proyecto: quién vota en Planning Poker y quién corrige sus horas.

La página **Crear cuenta** (`/registro`) existe y pide el email y la contraseña dos veces (entre 8 y 72 caracteres),
pero con el registro cerrado no deja crear cuentas nuevas.

### 1.2 Iniciar sesión

1. En **Iniciar sesión** completá tu **Email** y tu **Contraseña**.
2. Tocá **Iniciar sesión**.

Arriba a la derecha queda tu email, y desde ahí se entra a **Proyectos**.

Si algo falla, la página dice:
- "El email o la contraseña no son correctos.", sin decir cuál de los dos;
- "Demasiados intentos. Esperá un momento y volvé a probar.", después de varios intentos seguidos;
- "No pudimos conectar con el servicio de inicio de sesión. Probá de nuevo en unos minutos.", si Supabase no
  responde.

### 1.3 Salir y sesión vencida

- Para cerrar la sesión tocá **Salir**, arriba a la derecha. Vuelve a **Iniciar sesión** con el aviso "Cerraste
  sesión.".
- La sesión no se renueva sola: cuando vence, la app muestra "Tu sesión venció, iniciá sesión de nuevo." y hay que
  volver a ingresar. Cuánto dura lo define la configuración de Supabase.

### 1.4 Sin login (en tu computadora)

Si la app corre en tu computadora (`go run ./cmd/server` o `go run ./cmd/ejemplo`) o en un preview de Vercel, no pide
iniciar sesión. En las pantallas que necesitan saber quién sos, como Planning Poker o corregir horas, aparece un
desplegable para elegir el integrante.

## 2. Proyectos, sprints y reportes

### 2.1 Crear un proyecto

1. En **Proyectos** tocá **Nuevo proyecto** (o **Crear el primero** si todavía no hay ninguno).
2. Completá el **Nombre**, la **Descripción** (opcional) y las **fechas de inicio y de finalización**.
3. Tocá **Crear proyecto**.

Reglas:
- El nombre no se puede repetir y no importan las mayúsculas: "Metrics" y "metrics" cuentan como el mismo.
- La fecha de finalización tiene que ser posterior a la de inicio.
- Si algo está mal, la página marca todos los errores juntos, cada uno al lado de su campo, sin borrar lo que cargaste.

### 2.2 La vista del proyecto

Al entrar a un proyecto ves:
- su **estado**, que la aplicación calcula sola:
  - *Planificado* mientras no tenga sprints o no haya empezado el primero;
  - *En curso* desde que empieza el primer sprint;
  - *Finalizado* después de la fecha de fin;
- los **días restantes**, el **sprint activo** y los **story points completados**.

Desde ahí se llega a todo lo demás: **Backlog**, **Sprints**, **Integrantes**, **Horas**, **Registros de horas**,
**Defectos**, **Panel** y **Reporte del proyecto**.

Para cambiar los datos tocá **Modificar** y **Guardar cambios**. Valen las mismas reglas que al crear, y además:
- la fecha de fin no puede quedar antes del fin del último sprint;
- la fecha de inicio no puede quedar después del inicio del primero;
- un proyecto *Finalizado* ya no se modifica.

### 2.3 Registrar integrantes

1. En **Integrantes** tocá **Nuevo integrante**.
2. Completá el **Nombre**, el **Email** y el **Rol**: *Agile Enabler* o *Product Builder*.
3. Tocá **Registrar integrante**.

Reglas:
- El email no se repite dentro del proyecto.
- Hay un solo Agile Enabler activo por proyecto.
- El email tiene que ser el mismo con el que la persona entra a la aplicación. Así su cuenta queda vinculada a su
  integrante (sección 1).

Un integrante no se borra: con **Dar de baja** queda inactivo y se conserva todo lo que registró, como sus horas.

### 2.4 Crear un sprint y armar el Sprint Backlog

1. En **Sprints** tocá **Nuevo sprint**.
2. Completá el **Sprint Goal** (el objetivo, hasta 500 caracteres) y las fechas.
3. Tocá **Crear sprint**.

Reglas:
- El sprint tiene que quedar dentro de las fechas del proyecto.
- Uno nuevo empieza después de que termina el anterior.
- El número lo pone la aplicación: Sprint 1, Sprint 2, etc.

El **sprint activo** es el primero sin cerrar cuya fecha de inicio ya llegó.

En la vista del sprint, **Historias del backlog para asignar** muestra las historias listas para entrar al sprint.
Tocá **Asignar** en cada una que quieras sumar. Una historia está lista cuando:
- está estimada (sección 3);
- tiene al menos un criterio de aceptación;
- está *Pendiente* y no está en otro sprint.

Si los story points del sprint superan la velocidad del equipo, la página avisa, pero no lo impide. Con **Quitar**, una
historia vuelve al backlog mientras no haya empezado.

### 2.5 Durante el sprint y al cerrarlo

En el sprint activo, cada historia avanza con dos botones:
- **Empezar**: pasa a *En progreso*.
- **Completar**: pasa a *Hecho*. Para completarla tiene que tener todos sus criterios de aceptación cumplidos, que se
  marcan en la historia. La fecha de completado es la del día y no se deshace.

Para cerrar el sprint:
1. Tocá **Cerrar sprint**.
2. Confirmá en la página que sigue. Si todavía no llegó la fecha de fin, te avisa cuántos días faltan.

Al cerrar:
- Se guarda la **foto del cierre**: qué historias tenía el sprint, con sus story points, y cuáles se completaron. Esos
  valores quedan fijos aunque después se re-estime una historia.
- Las historias que no se completaron vuelven al backlog como *Pendientes*.
- Un sprint cerrado no se reabre ni se modifica.

La lista de **Sprints** muestra, para cada sprint cerrado, los SP completados y el **cumplimiento**: los SP completados
sobre los planificados. Al entrar a un sprint cerrado, la foto del cierre está en **Historias al cerrar**.

### 2.6 Reportes del sprint y del proyecto

**Reporte de un sprint:** en la vista de un sprint, tocá **Reporte del sprint**. Muestra:
- las métricas del sprint: SP planificados y completados, cumplimiento, historias completadas, velocidad, horas,
  desviación y defectos;
- la tabla de historias con sus horas;
- la tabla de defectos.

Si el sprint está cerrado, el reporte usa la foto del cierre. Si sigue abierto, muestra cómo va hasta hoy y lo avisa
arriba.

**Reporte del proyecto:** en la vista del proyecto, tocá **Reporte del proyecto**. Muestra:
- el resumen del proyecto, con los mismos valores del panel (sección 6.1);
- la evolución de la velocidad, sprint por sprint;
- un bloque por sprint, igual a su reporte.

### 2.7 Imprimir o descargar en PDF

Los dos reportes tienen dos botones:
- **Imprimir** usa la impresión del navegador. Se oculta la navegación y cada sprint arranca en una hoja nueva.
- **Descargar PDF** baja un archivo con el mismo contenido que la pantalla. El nombre lleva el proyecto, el sprint y la
  fecha, por ejemplo `software-metrics-sprint-2-2026-10-29.pdf`. El del proyecto no lleva el sprint:
  `software-metrics-2026-10-29.pdf`.

El PDF no puede mostrar emojis ni flechas, y en su lugar aparece un `?`. Las tildes y la ñ salen bien.

## 3. Backlog y estimación (Planning Poker)

### 3.1 Crear una historia

1. En la vista del proyecto tocá **Backlog** y después **Nueva historia** (o **Crear la primera** si todavía no hay
   ninguna).
2. Completá el **Título** (hasta 120 caracteres), la **Descripción** (opcional, hasta 2000) y la **Prioridad**: Alta,
   Media o Baja.
3. Si ya la estimaron, elegí los **Story points**. Si no, dejá *Sin estimar* y estimala más adelante (sección 3.5).
4. Tocá **Crear historia**.

El identificador (HU-1, HU-2…) lo pone la aplicación y la historia empieza **Pendiente**.

### 3.2 Criterios de aceptación

En la página de la historia, en **Criterios de aceptación**:
1. Escribí el criterio en **Nuevo criterio** (hasta 300 caracteres) y tocá **Agregar criterio**. Todo criterio empieza
   *No cumplido*.
2. Cuando se cumple, tocá **Marcar cumplido**. Si te equivocaste, **Marcar no cumplido**.

Arriba de la lista se ve el progreso, por ejemplo "2/3": dos de tres cumplidos. Los criterios importan después:
- una historia **sin criterios no se asigna** a un sprint;
- una historia se completa solo con **todos sus criterios cumplidos**.

Cuando la historia está Hecha, sus criterios ya no se cambian.

### 3.3 Editar o eliminar una historia

- **Editar:** en la página de la historia tocá **Editar**, cambiá el título, la descripción o la prioridad y tocá
  **Guardar cambios**. Se puede en cualquier estado. El identificador, el estado y los story points no cambian desde
  acá.
- **Eliminar:** tocá **Eliminar** y confirmá con **Eliminar historia**. No se puede eliminar una historia:
  - Hecha;
  - de un sprint cerrado;
  - con algo registrado: una sesión de Planning Poker, horas o defectos.

### 3.4 Ordenar, filtrar y buscar en el backlog

El **Backlog** muestra arriba las historias que faltan, ordenadas por prioridad (Alta, Media y Baja), y abajo las
**Hechas**. Cada fila muestra el estado, los story points y el progreso de los criterios.

- **Ordenar:** dentro de cada prioridad, **↑ Subir** y **↓ Bajar** mueven la historia un lugar. Así se decide qué va
  primero. Las Hechas no se reordenan.
- **Filtrar:** elegí un **Estado**, una **Prioridad** o escribí parte del título en **Buscar en el título**, y tocá
  **Filtrar**. Los filtros se combinan, y la búsqueda no distingue mayúsculas ni tildes. **Quitar filtros** vuelve a
  mostrar todo. Mientras haya filtros puestos no se puede reordenar.

### 3.5 Los estados de una historia

Una historia pasa por **Pendiente → En sprint → En progreso → Hecho**:

| Paso | Cómo |
|---|---|
| Pendiente → En sprint | Al asignarla a un sprint (sección 2.4). Necesita story points y criterios |
| En sprint → En progreso | En la página de la historia, **Empezar a trabajarla** |
| En progreso → Hecho | Al completarla desde el sprint activo, con todos los criterios cumplidos (sección 2.5) |
| En sprint o En progreso → Pendiente | Con **Quitar** en el sprint (sección 2.4), o al cerrar el sprint sin completarla (sección 2.5) |

Una historia Hecha queda así: no cambia de estado, no se re-estima, no se elimina y sus criterios quedan fijos.

### 3.6 Estimar con story points

En la página de la historia, en **Estimar**, elegí un valor de la escala **0, 1, 2, 3, 5, 8, 13 o 21** y tocá
**Guardar estimación**. Una historia Hecha ya no se estima.

Cada vez que cambia una estimación que ya tenía valor, queda anotada en el **Historial de estimaciones**: el valor
anterior, el nuevo y cuándo. El historial no se edita. Re-estimar una historia no cambia los sprints que ya se
cerraron, porque cada uno guarda sus story points al cerrarse.

### 3.7 Planning Poker

Sirve para que el equipo estime una historia votando sin ver lo que votan los demás. La sesión la conduce el **Agile
Enabler** del proyecto, y votan todos los integrantes activos.

**1. Abrir la sesión (Agile Enabler).** En la página de la historia tocá **Planning Poker** y después **Abrir
sesión**. Aparece la lista de **Participantes**: los integrantes activos del proyecto. Hay una sola sesión abierta por
historia, y una historia Hecha no se estima.

**2. Votar (todos).** Cada uno entra a la misma página, desde la compu o el celular, y toca la carta de su
estimación: 0, 1, 2, 3, 5, 8, 13 o 21. La página dice "Voto registrado. Podés cambiarlo hasta que se revele la
ronda.":
- **Nadie ve los valores**, ni siquiera el propio, hasta que se revela la ronda;
- debajo se ve **quién votó y quién falta** ("2 de 3 votaron"). Se actualiza sola cada 5 segundos;
- para cambiar el voto, se toca otra carta.

**3. Revelar.** Cuando votaron todos, la ronda **se revela sola**. Si alguien no va a votar, el Agile Enabler puede
tocar **Cerrar la ronda y revelar**, si hay al menos un voto. Aparece la tabla **Votos de la ronda** con el voto de
cada integrante, y las pantallas de los demás se actualizan solas.

**4. Registrar la estimación (Agile Enabler).** Después de conversar los votos, en **Registrar la estimación** elegí
el **Valor acordado** y tocá **Registrar estimación**. Con eso:
- la historia queda con ese valor de story points;
- si ya tenía uno distinto, el cambio va al historial de estimaciones (sección 3.6);
- la sesión se cierra y queda en **Sesiones cerradas**, con su valor acordado. Si hace falta, se puede abrir otra.

Si no hay acuerdo, el Agile Enabler registra el valor que se acuerde en la conversación, o abre una sesión nueva para
volver a votar.

Sin login (sección 1.4), en cada paso se elige en el desplegable quién abre, quién vota, quién cierra y quién
registra.

## 4. Tareas y esfuerzo

### 4.1 Dividir una historia en tareas

1. Entrá a la historia desde el **Backlog** y tocá **Tareas**.
2. En **Nueva tarea** completá el título, el responsable y las horas estimadas (con coma o punto, por ejemplo 1,5) y
   tocá **Agregar tarea**.
3. Arriba se ven las **horas estimadas de la historia**: la suma de sus tareas. Para cambiar las horas de una tarea usá
   **Cambiar horas**; cuando la terminás, marcala como completada.

### 4.2 Registrar las horas trabajadas

1. En la historia (o en sus tareas) tocá **Horas trabajadas**.
2. En **Registrar horas** elegí el integrante, la tarea (o "Toda la historia"), la fecha, qué hiciste y cuántas horas.
3. Tocá **Registrar horas**. Arriba se ven todos los registros de la historia y sus **horas reales**.

Reglas: entre 0 y 24 h por registro, y no más de 24 h por integrante en el mismo día. La fecha no puede ser futura ni
estar fuera de las fechas del proyecto. Solo registran horas los integrantes activos.

### 4.3 Consultar y corregir los registros de horas

1. En la vista del proyecto tocá **Registros de horas**.
2. Filtrá por integrante, historia o sprint y tocá **Filtrar**. Los filtros se combinan; el sprint toma los registros
   con fecha dentro del sprint. **Quitar filtros** vuelve a mostrar todo.
3. Abajo está la tabla **Horas estimadas vs reales por historia**, con la desviación de cada una.
4. Para corregir un registro tocá **Corregir**, elegí en **Quién corrige** tu nombre, cambiá lo que haga falta y tocá
   **Guardar corrección**. Cada integrante corrige solo sus registros; el Agile Enabler corrige los de todos.

### 4.4 Comparar horas estimadas y reales

En la vista del proyecto tocá **Horas**: muestra las horas estimadas, las reales y la desviación (en horas y en %) del
proyecto, de cada sprint y de cada historia. Una desviación positiva quiere decir que se tardó más de lo estimado. Si
no hay horas estimadas, el % dice "no aplicable".

## 5. Defectos

### 5.1 Registrar un defecto

1. En la vista del proyecto tocá **Defectos**.
2. En **Registrar defecto** completá la descripción, la severidad (Crítica, Alta, Media o Baja), el sprint en que se
   detectó y, si corresponde, la historia.
3. Tocá **Registrar defecto**. Todo defecto empieza **Abierto**.

### 5.2 Cambiar el estado

En la lista de **Defectos registrados**, cada defecto tiene botones con los estados a los que puede pasar: Abierto →
En progreso → Resuelto → Cerrado, y Reabierto si vuelve a aparecer. Al pasarlo a Resuelto hay que elegir el **sprint
en que se resolvió**. Un defecto reabierto deja de contar como resuelto.

### 5.3 Consultar los defectos

- **Filtros:** arriba de la lista, elegí severidad, estado y sprint de detección y tocá **Filtrar**. Se muestra la
  cantidad de defectos de cada severidad de la lista filtrada.
- **Resumen:** la sección **Resumen** muestra los defectos detectados y resueltos de cada sprint y del proyecto, y se
  puede filtrar por severidad.

## 6. Métricas, panel y gráficos

### 6.1 Panel del proyecto

En la vista del proyecto tocá **Panel**. Muestra seis tarjetas: **Estado** del proyecto y los días que quedan,
**Sprint activo**, **Velocidad** (promedio de los SP completados de los últimos 3 sprints cerrados), **% completado**
(SP completados sobre planificados), **Desviación de horas** y **Defectos abiertos**. La tarjeta que todavía no tiene
datos dice "sin datos".

### 6.2 Gráficos

En la vista del proyecto tocá **Gráficos**. Hay cinco:

| Gráfico | Qué muestra |
|---|---|
| Velocidad por sprint | Los SP completados de cada sprint cerrado y la velocidad hasta ese sprint |
| Story points planificados vs completados | Por sprint; un sprint cerrado usa los SP guardados al cerrarlo |
| Horas estimadas vs reales | Por sprint |
| Defectos por severidad | Los detectados y los que siguen abiertos |
| Burndown | Del sprint activo: los SP que faltan al final de cada día y la línea ideal. Con los botones de arriba se elige otro sprint |

Debajo de cada gráfico, **Ver los datos** abre una tabla con los mismos números. Los datos también salen en JSON en
`/proyectos/{id}/graficos/datos`.

### 6.3 Cómo se calculan las métricas

Las fórmulas, con un ejemplo completo, están en [`metricas.md`](metricas.md). Los sprints cerrados usan siempre la
"foto" que se guarda al cerrarlos, así las métricas de un sprint viejo no cambian aunque después se re-estime una
historia.
