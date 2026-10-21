# language: es
@HU-37
Característica: Reporte de sprint
  Como Agile Enabler
  Quiero un reporte de un sprint
  Para presentarlo en la Sprint Review

  Antecedentes:
    Dado el proyecto "Software Metrics" con estos sprints para el reporte:
      | sprint | goal                 | inicio     | fin        |
      | 1      | Gestionar el backlog | 2026-10-12 | 2026-10-18 |
      | 2      | Planificar sprints   | 2026-10-19 | 2026-10-25 |
    Y en el Sprint 1 estas historias con sus horas:
      | historia      | sp | completada | horas estimadas | horas reales |
      | HU-1 Login    | 5  | sí         | 6               | 8            |
      | HU-2 Registro | 3  | no         | 4               | 3            |
      | HU-3 Perfil   | 5  | no         | 0               | 0            |
    Y estos defectos:
      | defecto                        | severidad | historia | detectado en | resuelto en |
      | El login no valida el correo   | Crítica   | HU-1     | 1            | 1           |
      | El registro tarda mucho        | Media     | HU-2     | 1            |             |
    Y que el Sprint 1 se cerró para el reporte

  # Caso NORMAL: las métricas del sprint cerrado, calculadas con internal/metrics (CA-37.1 y CA-37.2)
  @CA-37.1 @CA-37.2
  Escenario: Las métricas del reporte de un sprint cerrado
    Cuando pido el reporte del Sprint 1
    Entonces el reporte muestra estas métricas:
      | métrica                        | valor                                |
      | SP planificados y completados  | 5 de 13 SP                           |
      | Cumplimiento                   | 38,5 %                               |
      | Historias completadas          | 1 de 3 (33,3 %)                      |
      | Velocidad del equipo           | 5,0 SP por sprint (1 sprint cerrado) |
      | Horas estimadas y reales       | 10 h estimadas · 11 h reales         |
      | Desviación                     | +1 h (+10,0 %)                       |
      | Defectos                       | 2 detectados · 1 resuelto            |

  # Caso NORMAL: las historias planificadas y completadas, con sus estimaciones y su esfuerzo (CA-37.1)
  @CA-37.1
  Escenario: Las historias del reporte con sus horas
    Cuando pido el reporte del Sprint 1
    Entonces el reporte muestra estas historias:
      | historia | título   | sp | resultado     | horas estimadas | horas reales | desviación         |
      | HU-1     | Login    | 5  | Completada    | 6 h             | 8 h          | +2 h (+33,3 %)     |
      | HU-2     | Registro | 3  | No completada | 4 h             | 3 h          | -1 h (-25,0 %)     |
      | HU-3     | Perfil   | 5  | No completada | 0 h             | 0 h          | 0 h (no aplicable) |
    Y el reporte muestra estos defectos:
      | defecto                      | severidad | estado   | detectado en | resuelto en |
      | El login no valida el correo | Crítica   | Resuelto | Sprint 1     | Sprint 1    |
      | El registro tarda mucho      | Media     | Abierto  | Sprint 1     | —           |

  # Caso ALTERNATIVO: el reporte de un sprint cerrado no cambia con lo que pasa después (RN1)
  @CA-37.1
  Escenario: El reporte conserva la foto del cierre
    Dado que "HU-2" se re-estimó en 8 story points y entró al Sprint 2 del reporte
    Cuando pido el reporte del Sprint 1
    Entonces el reporte muestra la métrica "SP planificados y completados" con "5 de 13 SP"
    Y el reporte muestra "HU-2" con 3 SP y "No completada"

  # Caso ALTERNATIVO: un sprint abierto muestra cómo va y avisa que es parcial (RN2 y RN4)
  @CA-37.1
  Escenario: Reporte de un sprint que todavía no se cerró
    Dado que "HU-2" se re-estimó en 8 story points y entró al Sprint 2 del reporte
    Y el defecto "El registro tarda mucho" se resolvió en el Sprint 2
    Cuando pido el reporte del Sprint 2
    Entonces el reporte avisa "El sprint no está cerrado: el reporte muestra cómo va hasta hoy."
    Y el reporte muestra la métrica "SP planificados y completados" con "0 de 8 SP"
    Y el reporte muestra la métrica "Velocidad del equipo" con "5,0 SP por sprint (1 sprint cerrado)"
    Y el reporte muestra la métrica "Defectos" con "0 detectados · 1 resuelto"

  # Caso LÍMITE: un sprint sin historias ni defectos no divide por cero (RN3, RN5 y RN6)
  @CA-37.1
  Escenario: Reporte de un sprint sin historias ni defectos
    Cuando pido el reporte del Sprint 2
    Entonces el reporte muestra la métrica "Historias completadas" con "0 de 0 (0,0 %)"
    Y el reporte muestra la métrica "Desviación" con "0 h (no aplicable)"
    Y el reporte dice "No hubo defectos en este sprint."

  # Caso de ERROR: el sprint no existe (CA-37.3)
  @CA-37.3
  Escenario: Pedir el reporte de un sprint que no existe
    Cuando pido el reporte del Sprint 9
    Entonces el reporte responde 404 con "No encontramos ese sprint."
