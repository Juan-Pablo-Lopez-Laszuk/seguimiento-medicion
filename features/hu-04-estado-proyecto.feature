# language: es
@HU-04
Característica: Consultar estado del proyecto
  Como integrante
  Quiero consultar el estado del proyecto
  Para saber cómo avanza

  Antecedentes:
    Dado el proyecto "Software Metrics" del "2026-10-05" al "2026-11-01" para consultar su estado

  # Caso LÍMITE: proyecto sin sprints, aunque ya hayan pasado las fechas (CA-04.3)
  @CA-04.3
  Esquema del escenario: Un proyecto sin sprints está Planificado
    Dado que hoy es "<hoy>"
    Cuando consulto el estado del proyecto
    Entonces el estado es "Planificado"
    Y no hay sprint activo

    Ejemplos:
      | hoy        |
      | 2026-10-20 |
      | 2026-11-10 |

  # Caso NORMAL y LÍMITE: el estado según la fecha (CA-04.2)
  @CA-04.2
  Esquema del escenario: El estado depende de la fecha y los sprints
    Dado que el proyecto tiene el Sprint 1 del "2026-10-05" al "2026-10-11"
    Y que hoy es "<hoy>"
    Cuando consulto el estado del proyecto
    Entonces el estado es "<estado>"
    Y al proyecto le quedan <dias> días

    Ejemplos:
      | hoy        | estado      | dias |
      | 2026-10-04 | Planificado | 28   |
      | 2026-10-05 | En curso    | 27   |
      | 2026-11-01 | En curso    | 0    |
      | 2026-11-02 | Finalizado  | 0    |

  # Caso NORMAL: sprint activo y días que le quedan (CA-04.1)
  @CA-04.1
  Escenario: Proyecto en curso con su sprint activo
    Dado que el proyecto tiene el Sprint 1 del "2026-10-05" al "2026-10-11"
    Y que el proyecto tiene el Sprint 2 del "2026-10-12" al "2026-10-18"
    Y que el Sprint 1 está cerrado
    Y que hoy es "2026-10-13"
    Cuando consulto el estado del proyecto
    Entonces el estado es "En curso"
    Y el sprint activo es el Sprint 2 y le quedan 5 días

  # Caso ALTERNATIVO: el sprint terminó por fecha pero no se cerró; sigue activo (CA-04.1)
  @CA-04.1
  Escenario: Un sprint terminado sin cerrar sigue activo
    Dado que el proyecto tiene el Sprint 1 del "2026-10-05" al "2026-10-11"
    Y que el proyecto tiene el Sprint 2 del "2026-10-12" al "2026-10-18"
    Y que hoy es "2026-10-13"
    Cuando consulto el estado del proyecto
    Entonces el sprint activo es el Sprint 1 y le quedan 0 días
    Y se avisa que el sprint terminó y falta cerrarlo

  # Caso NORMAL y de ERROR controlado: porcentaje de SP completados (CA-04.1)
  @CA-04.1
  Esquema del escenario: Porcentaje de story points completados
    Dado que el proyecto tiene historias con <hechos> SP Hechos de <estimados> SP estimados
    Cuando consulto el estado del proyecto
    Entonces el porcentaje de SP completados es "<porcentaje>"

    Ejemplos:
      | hechos | estimados | porcentaje              |
      | 3      | 8         | 37,5 %                  |
      | 8      | 8         | 100,0 %                 |
      | 0      | 0         | Sin historias estimadas |
