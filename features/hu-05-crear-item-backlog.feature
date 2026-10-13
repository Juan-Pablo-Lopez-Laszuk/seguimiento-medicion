# language: es
@HU-05
Característica: Crear ítem del Product Backlog
  Como Product Builder
  Quiero agregar historias al backlog con título, descripción, prioridad y story points opcionales
  Para registrar el trabajo pendiente

  Antecedentes:
    Dado el proyecto "Software Metrics" para el backlog

  # Caso NORMAL (CA-05.1, CA-05.2 y CA-05.3)
  @CA-05.1 @CA-05.2 @CA-05.3
  Escenario: Crear la primera historia del backlog
    Cuando creo la historia "Iniciar sesión con Google" con prioridad "Alta" sin story points
    Entonces se crea la historia HU-1 "Iniciar sesión con Google"
    Y la historia queda en estado "Pendiente"
    Y la historia queda sin estimar

  # Caso ALTERNATIVO: los story points son opcionales al crear (CA-05.3)
  @CA-05.3
  Escenario: Crear una historia ya estimada
    Cuando creo la historia "Ver el tablero" con prioridad "Media" y 5 story points
    Entonces se crea la historia HU-1 "Ver el tablero"
    Y la historia queda con 5 story points

  # Caso ALTERNATIVO: el identificador sigue al último (CA-05.1)
  @CA-05.1
  Escenario: La siguiente historia lleva el número que sigue
    Dado que el backlog ya tiene la historia "Primera" con prioridad "Alta"
    Cuando creo la historia "Segunda" con prioridad "Baja" sin story points
    Entonces se crea la historia HU-2 "Segunda"

  # Caso ALTERNATIVO: cada proyecto numera sus propias historias (CA-05.1)
  @CA-05.1
  Escenario: Cada proyecto numera sus propias historias
    Dado que el backlog ya tiene la historia "Primera" con prioridad "Alta"
    Y el proyecto "Otro proyecto" para el backlog
    Cuando creo en "Otro proyecto" la historia "Arranque" con prioridad "Media" sin story points
    Entonces se crea la historia HU-1 "Arranque"

  # Caso LÍMITE: las tres prioridades valen; cualquier otra, o ninguna, no (CA-05.2)
  @CA-05.2
  Esquema del escenario: Prioridad de la historia
    Cuando creo la historia "Prioridad" con prioridad "<prioridad>" sin story points
    Entonces la historia da como resultado "<resultado>"

    Ejemplos:
      | prioridad | resultado                                                |
      | Alta      | creada                                                   |
      | Media     | creada                                                   |
      | Baja      | creada                                                   |
      | Urgente   | prioridad: la prioridad tiene que ser Alta, Media o Baja |
      |           | prioridad: la prioridad tiene que ser Alta, Media o Baja |

  # Caso LÍMITE: los extremos de la escala valen; fuera de ella, no (CA-05.3)
  @CA-05.3
  Esquema del escenario: Story points de la historia
    Cuando creo la historia "Estimada" con prioridad "Media" y "<puntos>" como story points
    Entonces la historia da como resultado "<resultado>"

    Ejemplos:
      | puntos | resultado                                                                         |
      | 0      | creada                                                                            |
      | 21     | creada                                                                            |
      | 4      | story_points: los story points tienen que ser 0, 1, 2, 3, 5, 8, 13 o 21          |
      | 22     | story_points: los story points tienen que ser 0, 1, 2, 3, 5, 8, 13 o 21          |
      | -1     | story_points: los story points tienen que ser 0, 1, 2, 3, 5, 8, 13 o 21          |
      | abc    | story_points: los story points tienen que ser 0, 1, 2, 3, 5, 8, 13 o 21          |

  # Caso LÍMITE: el título tiene hasta 120 caracteres (CA-05.2)
  @CA-05.2
  Esquema del escenario: Largo del título
    Cuando creo una historia con un título de <largo> caracteres
    Entonces la historia da como resultado "<resultado>"

    Ejemplos:
      | largo | resultado                                                 |
      | 120   | creada                                                    |
      | 121   | titulo: el título no puede superar los 120 caracteres     |

  # Caso de ERROR: título vacío; todos los errores se informan juntos (CA-05.2)
  @CA-05.2
  Escenario: El título es obligatorio
    Cuando creo la historia "   " con prioridad "" sin story points
    Entonces la historia da como resultado "titulo: el título es obligatorio"
    Y la historia da como resultado "prioridad: la prioridad tiene que ser Alta, Media o Baja"
    Y el backlog no tiene historias
