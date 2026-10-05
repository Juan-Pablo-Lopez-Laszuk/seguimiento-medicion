# Sprint Review · Sprint 0 · 05/10/2026

Participantes: López, Piastrellini, Bravo · Product Architect (profesores): por confirmar

## Sprint Goal

> Dejar listo el entorno de trabajo, el tablero y el Product Backlog inicial para empezar a construir en el Sprint 1.
> ¿Se cumplió? **Parcialmente**: tablero, backlog, plantillas, métricas, modelo de datos y deploy listos; el esqueleto
> en Go (TEC-01) quedó en revisión y la integración continua (TEC-02) no se empezó.

## Terminado

| Issue | Tarea | SP | Responsable | Evidencia |
|---|---|---|---|---|
| #32 | Definición y fórmula de las métricas | 3 | Carolina | `docs/metricas.md` (PR #31) |
| #33 | Plantilla BDD y escenarios de HU-31 | 2 | Carolina | `features/` (PR #36) |
| #34 | PR de práctica | 1 | Carolina | PR #35 |
| #54 | Vercel: prueba hola mundo | 2 | Carolina | https://seguimiento-medicion.vercel.app (PR #55) |
| #59 | Registro de uso de IA | 1 | Carolina | `docs/ia/registro.md` (PR #60) |
| #61 | Spec SDD de HU-31 | 2 | Carolina | `specs/HU-31-velocidad.md` (PR #62) |
| #64 | Propuesta de modelo de datos | — | Mariano | PR #56 |
| #65 | Plantilla de especificación SDD | — | Mariano | `specs/_plantilla.md` (PR #57) |
| #66 | Spike Supabase Auth (JWT ES256) | — | Mariano | PR #58 |
| #51 | Tablero, etiquetas, hitos y plantillas | — | Juan Pablo | GitHub Project con vistas, campos y automatizaciones |
| #52 | Preguntas abiertas con los profesores | — | Juan Pablo | Respuestas en la sección siguiente |

## No terminado (pasa al Sprint 1)

| Issue | Tarea | SP | Motivo |
|---|---|---|---|
| #37 | TEC-01 · Esqueleto del proyecto | 3 | En revisión: el preview de Vercel falló al importar `internal/`; corregido en el PR #73 |
| #71 | godog con los escenarios de HU-31 | 2 | En revisión (PR #72), depende de TEC-01 |
| #38 | TEC-02 · Integración continua | 2 | No se empezó; depende de TEC-01 |

## Preguntas abiertas (#52)

Las que dependen de la cátedra se consultaron a los profesores; las que la consigna deja abiertas las decidió el equipo.

| Pregunta | Quién la resolvió | Decisión |
|---|---|---|
| Fecha de la presentación final | Profesores | Primeras semanas de noviembre (el Sprint 4 cierra el 01/11) |
| Cobertura mínima de tests | Profesores | No hay un porcentaje fijo: hay que cubrir con tests todo lo que pide la consigna |
| Formato de la documentación final | Profesores | Markdown en el repositorio |
| Escala de story points | Equipo | Fibonacci 0, 1, 2, 3, 5, 8, 13, 21 y "?" |
| "Diferencia" en Planning Poker | Equipo | Hay diferencia si los votos extremos están a más de una posición de la escala |
| Horas estimadas: ¿por historia, por tarea o ambas? | Equipo | Ambas: la historia suma las horas de sus tareas (la consigna dice "historia o tarea") |
| Sprints para la velocidad | Equipo | Promedio de los últimos 3 sprints cerrados |

## Feedback del cliente

- Respuestas a las preguntas abiertas (tabla de arriba).
- (agregar cualquier otro comentario de los profesores)

## Cambios en el Product Backlog

- TEC-01, TEC-02 y la tarea de godog de HU-31 pasan al Sprint 1.
- Cobertura: cada requerimiento de la consigna tiene que tener al menos un test o escenario BDD que lo pruebe.
- Documentación final en Markdown (el reporte PDF que genera la app, HU de la E9, sigue siendo obligatorio).
- Nuevo aprendizaje técnico: la función de Vercel no puede importar paquetes `internal/`; se entra por el paquete público `app/`.
