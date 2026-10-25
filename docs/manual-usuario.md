# Manual de usuario

Entregable 12 de la consigna. Una sección por épica, con los pasos para usar cada pantalla de la aplicación.

> Secciones 1 y 3: Mariano. Sección 2: Juan Pablo. Secciones 4 a 6: Carolina.

## 1. Ingresar a la aplicación

> Sección a cargo de Mariano (registro, ingreso y salida).

## 2. Proyectos, sprints y reportes

> Sección a cargo de Juan Pablo.

## 3. Backlog y estimación (Planning Poker)

> Sección a cargo de Mariano.

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
