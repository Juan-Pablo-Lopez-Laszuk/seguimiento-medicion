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

## No terminado (pasa al Sprint 1)

| Issue | Tarea | SP | Motivo |
|---|---|---|---|
| #37 | TEC-01 · Esqueleto del proyecto | 3 | En revisión: el preview de Vercel falló al importar `internal/`; corregido en el PR #73 |
| #71 | godog con los escenarios de HU-31 | 2 | En revisión (PR #72), depende de TEC-01 |
| #38 | TEC-02 · Integración continua | 2 | No se empezó; depende de TEC-01 |
| #52 | Preguntas abiertas a los profesores | — | Falta la respuesta de la cátedra (ver abajo) |

## Preguntas abiertas (#52)

Propuesta del equipo; se mantiene salvo que los profesores digan otra cosa.

| Pregunta | Propuesta | Respuesta de la cátedra |
|---|---|---|
| Fecha exacta de la presentación final | Semana del 02/11 (cierre del Sprint 4 el 01/11) | |
| Escala de story points | Fibonacci 0, 1, 2, 3, 5, 8, 13, 21 y "?" | |
| "Diferencia" en Planning Poker | Hay diferencia si los votos extremos están a más de una posición de la escala | |
| Horas estimadas: ¿por historia, por tarea o ambas? | Ambas: la historia suma las horas de sus tareas | |
| Sprints para la velocidad | Promedio de los últimos 3 sprints cerrados | |
| Cobertura mínima de tests | 80 % en `internal/domain` y `internal/metrics` | |
| Formato de la documentación final | Markdown en el repositorio + PDF exportado | |

## Feedback del cliente

- (completar con lo que digan los profesores)

## Cambios en el Product Backlog

- TEC-01, TEC-02 y la tarea de godog de HU-31 pasan al Sprint 1.
- Nuevo aprendizaje técnico: la función de Vercel no puede importar paquetes `internal/`; se entra por el paquete público `app/`.
