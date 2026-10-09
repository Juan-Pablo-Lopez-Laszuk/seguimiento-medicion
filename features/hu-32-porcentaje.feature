# language: es
@HU-32
Característica: Porcentaje de historias completadas
  Como Agile Enabler
  Quiero ver qué porcentaje de las historias planificadas se completó
  Para saber cuánto del compromiso del sprint se cumplió

  # Caso NORMAL (CA-32.1)
  @CA-32.1
  Escenario: Porcentaje de un sprint
    Dado un sprint con 4 historias de las cuales 3 están hechas
    Cuando consulto el porcentaje de completadas del sprint
    Entonces el porcentaje de historias completadas es "75,0 %"

  # Caso LÍMITE (CA-32.1): todas o ninguna
  @CA-32.1
  Esquema del escenario: Sprint con todas o ninguna historia hecha
    Dado un sprint con <planificadas> historias de las cuales <hechas> están hechas
    Cuando consulto el porcentaje de completadas del sprint
    Entonces el porcentaje de historias completadas es "<porcentaje>"

    Ejemplos:
      | planificadas | hechas | porcentaje |
      | 3            | 3      | 100,0 %    |
      | 5            | 0      | 0,0 %      |

  # Caso de ERROR controlado (CA-32.2): sin historias no se divide por cero
  @CA-32.2
  Escenario: Sprint sin historias
    Dado un sprint sin historias planificadas
    Cuando consulto el porcentaje de completadas del sprint
    Entonces el porcentaje de historias completadas es "0,0 %"

  # Caso ALTERNATIVO (CA-32.3)
  @CA-32.3
  Escenario: Porcentaje del proyecto con varios sprints
    Dado un sprint con 4 historias de las cuales 3 están hechas
    Y otro sprint con 3 historias de las cuales 3 están hechas
    Cuando consulto el porcentaje de completadas del proyecto
    Entonces el porcentaje de historias completadas es "85,7 %"
