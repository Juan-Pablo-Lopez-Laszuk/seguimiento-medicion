# language: es
@HU-23
Característica: Registrar horas estimadas
  Como Product Builder
  Quiero cargar las horas estimadas de una historia
  Para después compararlas con las horas reales

  # Caso NORMAL (CA-23.1)
  @CA-23.1
  Escenario: Cargar horas estimadas en una historia sin tareas
    Cuando cargo "8" horas estimadas en una historia sin tareas
    Entonces las horas estimadas de la historia son "8"

  # Caso LÍMITE (CA-23.1): se aceptan decimales, con coma o con punto
  @CA-23.1
  Esquema del escenario: Horas con decimales
    Cuando cargo "<horas>" horas estimadas en una historia sin tareas
    Entonces las horas estimadas de la historia son "<resultado>"

    Ejemplos:
      | horas | resultado |
      | 1,5   | 1,5       |
      | 0,25  | 0,25      |
      | 2.5   | 2,5       |

  # Caso de ERROR (CA-23.1): las horas tienen que ser mayores a 0
  @CA-23.1
  Esquema del escenario: Horas inválidas
    Cuando cargo "<horas>" horas estimadas en una historia sin tareas
    Entonces se rechaza la estimación con el error "las horas estimadas deben ser mayores a 0"

    Ejemplos:
      | horas |
      | 0     |
      | -2    |
      | abc   |

  # Caso ALTERNATIVO (CA-23.2): si la historia tiene tareas, se suman las horas de las tareas
  @CA-23.2
  Escenario: Historia con tareas
    Dado una historia estimada en "10" horas
    Y una tarea estimada en "2" horas
    Y una tarea estimada en "3,5" horas
    Cuando consulto las horas estimadas de la historia
    Entonces las horas estimadas de la historia son "5,5"
