# language: es
@HU-38
Característica: Reporte de proyecto
  Como Agile Enabler
  Quiero un reporte consolidado del proyecto
  Para rendir cuentas al cliente

  Antecedentes:
    Dado el proyecto "Software Metrics" con estos sprints para el reporte del proyecto:
      | sprint | goal                 | inicio     | fin        |
      | 1      | Gestionar el backlog | 2026-10-12 | 2026-10-18 |
      | 2      | Planificar sprints   | 2026-10-19 | 2026-10-25 |
      | 3      | Ver métricas         | 2026-10-26 | 2026-11-01 |
    Y estas historias del proyecto con sus horas:
      | historia      | sp | horas estimadas | horas reales |
      | HU-1 Login    | 5  | 6               | 8            |
      | HU-2 Registro | 3  | 4               | 3            |
      | HU-3 Perfil   | 5  | 2               | 0            |
    Y que el Sprint 1 se cerró con "HU-1, HU-2, HU-3" y completó "HU-1"
    Y que el Sprint 2 se cerró con "HU-2, HU-3" y completó "HU-2"
    Y que "HU-3" está en el Sprint 3
    Y estos defectos del proyecto:
      | defecto                      | severidad | historia | detectado en | resuelto en |
      | El login no valida el correo | Crítica   | HU-1     | 1            | 1           |
      | El registro tarda mucho      | Media     | HU-2     | 1            | 2           |
      | El perfil no carga la foto   | Baja      |          | 2            |             |

  # Caso NORMAL: el resumen del proyecto, con los totales del proyecto y no la suma de los sprints (CA-38.1, RN2)
  @CA-38.1
  Escenario: El resumen del reporte del proyecto
    Cuando pido el reporte del proyecto
    Entonces el reporte del proyecto muestra este resumen:
      | métrica                  | valor                                  |
      | Sprints                  | 2 cerrados de 3                        |
      | SP completados           | 8 de 13 SP (61,5 %)                    |
      | Velocidad del equipo     | 4,0 SP por sprint (2 sprints cerrados) |
      | Horas estimadas y reales | 12 h estimadas · 11 h reales           |
      | Desviación               | -1 h (-8,3 %)                          |
      | Defectos                 | 3 detectados · 2 resueltos · 1 abierto |

  # Caso NORMAL: cada sprint con el mismo contenido que su reporte de HU-37 (CA-38.1, RN1)
  @CA-38.1
  Escenario: Un bloque por sprint igual a su reporte
    Cuando pido el reporte del proyecto
    Entonces el reporte del proyecto tiene los bloques "Sprint 1, Sprint 2, Sprint 3"
    Y cada bloque muestra lo mismo que el reporte de ese sprint
    Y el bloque del "Sprint 3" avisa "El sprint no está cerrado: el reporte muestra cómo va hasta hoy."

  # Caso NORMAL: la evolución de la velocidad, solo con los sprints cerrados (CA-38.2, RN3)
  @CA-38.2
  Escenario: La evolución de la velocidad
    Cuando pido el reporte del proyecto
    Entonces el reporte del proyecto muestra esta evolución de la velocidad:
      | sprint   | sp planificados | sp completados | cumplimiento | velocidad         |
      | Sprint 1 | 13              | 5              | 38,5 %       | 5,0 SP por sprint |
      | Sprint 2 | 8               | 3              | 37,5 %       | 4,0 SP por sprint |

  # Caso ALTERNATIVO: lo que pasa después del cierre no cambia el reporte (RN1 y RN3)
  @CA-38.1 @CA-38.2
  Escenario: El reporte del proyecto conserva las fotos de los cierres
    Dado que "HU-3" se re-estimó en 8 story points para el reporte del proyecto
    Cuando pido el reporte del proyecto
    Entonces la fila del "Sprint 2" en la evolución tiene 8 SP planificados y 3 completados
    Y el bloque del "Sprint 2" muestra "HU-3" con 5 SP

  # Caso LÍMITE: un proyecto sin sprints no da error (RN5)
  @CA-38.1 @CA-38.2
  Escenario: Reporte de un proyecto sin sprints
    Dado el proyecto "Proyecto nuevo" sin sprints
    Cuando pido el reporte del proyecto "Proyecto nuevo"
    Entonces la evolución de la velocidad dice "Aún no hay sprints cerrados"
    Y el reporte del proyecto dice "Todavía no hay sprints en este proyecto."

  # Caso de ERROR: el proyecto no existe
  @CA-38.1
  Escenario: Pedir el reporte de un proyecto que no existe
    Cuando pido el reporte del proyecto 99
    Entonces el reporte del proyecto responde 404 con "No encontramos ese proyecto."
