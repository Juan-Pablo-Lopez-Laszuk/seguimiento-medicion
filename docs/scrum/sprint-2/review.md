# Sprint Review · Sprint 2 · 18/10/2026

Participantes: López, Piastrellini, Bravo · Product Architect (profesores): [A COMPLETAR]

## Sprint Goal

> Trabajar el backlog y el sprint desde la app: ver el estado del proyecto, completar las historias (criterios,
> estimación y estado), asignarlas al sprint y marcarlas como completadas, todo con un diseño base común; y empezar a
> registrar tareas, esfuerzo y defectos.

¿Se cumplió? **Sí.** Desde la app ya se ve el estado del proyecto, se completa una historia (criterios, estimación,
estado, edición y eliminación), se asigna al sprint y se marca como completada, todo con el diseño base de TEC-07. Y
ya se registran tareas (HU-25), esfuerzo (HU-24) y defectos (HU-27). HU-24 y HU-27 llegaron a revisión el sábado 17/10
a la noche y entraron el domingo 18/10 a la noche, el último día del sprint.

## Terminado

| Issue | Historia / tarea | SP | Responsable | Evidencia |
|---|---|---|---|---|
| #42 | HU-04 · Consultar estado del proyecto | 3 | Juan Pablo | PR #97 · `/proyectos/{id}` |
| #40 | HU-02 · Modificar proyecto | 2 | Juan Pablo | PR #98 · `/proyectos/{id}/editar` |
| #16 | TEC-07 · Diseño base de la interfaz | 3 | Carolina | PR #99 · `web/templates/base.html`; lo pendiente de formularios y tablas quedó en #101 |
| #20 | HU-07 · Gestionar criterios de aceptación | 2 | Mariano | PR #100 · `/proyectos/{id}/backlog/{numero}` |
| #3 | HU-25 · Descomponer historia en tareas | 3 | Carolina | PR #102 · `/proyectos/{id}/backlog/{numero}/tareas` |
| #21 | HU-08 · Cambiar el estado de una historia | 3 | Mariano | PR #103 · estado en la página de la historia y un color por estado en el backlog |
| #44 | HU-12 · Asignar historias a un sprint | 5 | Juan Pablo | PR #105 · `/proyectos/{id}/sprints/{numero}` |
| #45 | HU-13 · Registrar historias completadas | 2 | Juan Pablo | PR #107 · botones Empezar y Completar en la página del sprint activo |
| #24 | HU-16 · Estimar historia con story points (sin el historial) | 2 | Mariano | PR #108 · formulario de estimar en la página de la historia; el historial pasó a HU-16b (#106) |
| #19 | HU-06 · Editar y eliminar ítem | 2 | Mariano | PR #109 · `/proyectos/{id}/backlog/{numero}/editar` y `/eliminar` |
| #2 | HU-24 · Registrar esfuerzo realizado | 3 | Carolina | PR #110 · `/proyectos/{id}/backlog/{numero}/esfuerzo` (18/10) |
| #5 | HU-27 · Registrar defecto | 3 | Carolina | PR #111 · `/proyectos/{id}/defectos` (18/10) |

Fuera del compromiso: workflow que pide la revisión según la rotación al abrir un PR (#104, Juan Pablo, sin SP).

**Total terminado: 33 SP de 33 comprometidos** (27 al sábado 17/10 y 6 el domingo 18/10). Por integrante: Juan Pablo
12 de 12 · Mariano 9 de 9 · Carolina 12 de 12.

## Demo

Lo que se muestra en la app:

1. **Diseño base** (TEC-07): menú que se adapta al celular, mensajes y componentes comunes en todas las pantallas.
2. **Estado del proyecto** (HU-04): Planificado, En curso o Finalizado, el sprint activo, los días que quedan y el % de
   story points completados. **Modificar proyecto**
   (HU-02) con las mismas validaciones del alta.
3. **Historia completa** (HU-07, HU-16, HU-08 y HU-06): criterios de aceptación con su progreso, estimación con la
   escala Fibonacci, cambio de estado, edición, y eliminación con una pantalla que avisa cuántos criterios y tareas se
   pierden. Una historia Hecha no se re-estima ni se elimina.
4. **Tareas de una historia** (HU-25): alta, estimación en horas y marcar completada.
5. **Sprint Backlog** (HU-12 y HU-13): asignar historias listas (estimadas y con criterios) al sprint, con el motivo de
   las que no se pueden y el aviso cuando el sprint supera la velocidad del equipo; en el sprint activo, empezar y
   completar historias (solo con todos los criterios cumplidos), con la fecha de completado y el avance del sprint.
6. **Esfuerzo** (HU-24): horas trabajadas por integrante en una historia o en una de sus tareas, con el tope de 24 h
   por día y las horas reales de la historia arriba. **Defectos** (HU-27): página de defectos del proyecto, con
   severidad, sprint de detección e historia opcional; todo defecto arranca Abierto.

## Feedback del cliente

- [A COMPLETAR en la Review]

## Cambios en el Product Backlog

- **HU-16 se dividió (decisión de Juan Pablo del 15/10):** CA-16.3 (re-estimar conserva el valor anterior en un
  historial) necesita una tabla nueva y la migración 0003, así que pasó a **HU-16b (#106, 2 SP)**. HU-16 quedó con
  CA-16.1 y CA-16.2, sin migración.
- **HU-06 cambia una regla de HU-05:** si se elimina la última historia, su número queda libre para la próxima (quedó
  escrito en la spec de HU-05).
- **#101** (lo que falta de formularios y tablas de TEC-07) queda en el Product Backlog.
- **Para el Sprint 3**, además de lo que la Planning ya pasaba (HU-14, HU-15, HU-09, HU-10, HU-17 a HU-19 y TEC-05):
  HU-16b (#106) y HU-28 (gestionar el estado de un defecto), que ya tiene HU-27 en `main`. **HU-14 (cerrar sprint)** es la primera: con HU-12 y HU-13
  en `main` ya no le falta nada, y sin ella la velocidad no se puede calcular desde la app. Al hacerla hay que actualizar
  `EnSprintCerrado` del repositorio en memoria, que hoy devuelve siempre `false` porque todavía no se puede cerrar un
  sprint.
