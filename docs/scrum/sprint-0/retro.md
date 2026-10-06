# Retrospectiva · Sprint 0 · 05/10/2026

Participantes: López, Piastrellini, Bravo · Facilitador: Juan Pablo

> Borrador: los puntos son una propuesta a partir de lo que pasó en el repo. Cada uno agrega, saca o corrige en la reunión.

## Métricas del sprint

SP comprometidos al inicio: 18 · SP completados al cierre: 21 · Velocidad: 21 (referencia; el Sprint 0 fue de preparación) · Defectos abiertos: 0

Las tareas de Mariano (#64, #65, #66) y de Juan Pablo (#51, #52, #53) se cargaron sin story points y se estimaron
**después de terminadas**, para que el trabajo del Sprint 0 también quede medido:

| Integrante | Tareas | SP estimados después |
|---|---|---|
| Mariano | #64 (3), #65 (1), #66 (2) | 6 |
| Juan Pablo | #51 (3), #52 (1) | 4 |
| Juan Pablo | #53 (2), no terminada al cierre | — |

Como son estimaciones a posteriori, no estaban en los 18 SP comprometidos al inicio: sin ellas, los completados serían 11
(todos de Carolina, la única que cargó los SP antes de empezar).

## ¿Qué salió bien?

- Todos hicieron su primer Pull Request y revisión siguiendo la rotación (JP → Mariano → Carolina → JP).
- El backlog completo quedó cargado en el tablero con SP, prioridad, épica y sprint.
- Carolina encontró y corrigió el error de Vercel en el esqueleto (PR #73) antes de que llegara a `main`.
- Se definieron métricas, modelo de datos, plantilla SDD y login con Supabase antes de escribir código.

## ¿Qué podemos mejorar?

- TEC-01 recién estuvo para revisar el jueves 01/10 a la noche y se mergeó el lunes 05/10: bloqueó el trabajo de los
  demás en Go. Lo que bloquea a otros va primero.
- Se dio por probado algo que solo se probó en local: el deploy en Vercel compila distinto.
- Seis tareas no tenían story points al empezar (las de Mariano y las de Juan Pablo se estimaron recién al terminarlas)
  y no se pueden medir con precisión. Le pasó también al Agile Enabler, que es quien tiene que controlarlo.
- No hubo dailies registradas.

## Acciones

| Acción | Responsable | Para cuándo |
|---|---|---|
| Mergear TEC-01 y armar el CI (TEC-02) antes de empezar las historias | Juan Pablo | Martes 06/10 (hecho el 05/10) |
| Todo issue del sprint lleva story points antes de la Planning; al cerrarla, el Agile Enabler revisa la vista Sprint actual y no arranca el sprint con issues sin SP | Todos · controla Juan Pablo | Planning de cada lunes |
| Un PR no se aprueba sin el preview de Vercel en verde | Todos | Desde el Sprint 1 |
| Daily por mensaje (qué hice / qué hago / bloqueos) en el issue del sprint | Todos | Lunes a viernes |

## Seguimiento de las acciones del sprint anterior

- No aplica (primer sprint).
