# language: es
@HU-16
Característica: Estimar historia con story points
  Como equipo
  Queremos asignar story points a una historia
  Para dimensionar su esfuerzo relativo

  Antecedentes:
    Dado el proyecto "Software Metrics" con la historia "Iniciar sesión con Google" para estimar

  # Caso NORMAL (CA-16.1): una historia sin estimar recibe su primera estimación
  @CA-16.1
  Escenario: Estimar una historia sin estimar
    Cuando estimo la historia en "5" story points
    Entonces la estimación de la historia es "5"

  # Caso ALTERNATIVO (CA-16.1): re-estimar reemplaza el valor anterior (el historial es HU-16b)
  @CA-16.1
  Escenario: Re-estimar reemplaza el valor anterior
    Dado que la historia ya está estimada en "5" story points
    Cuando estimo la historia en "8" story points
    Entonces la estimación de la historia es "8"

  # Caso LÍMITE (RN5): pedir el mismo valor no es un error
  @CA-16.1
  Escenario: Estimar con el mismo valor que ya tiene
    Dado que la historia ya está estimada en "5" story points
    Cuando estimo la historia en "5" story points
    Entonces la estimación da como resultado "ok"
    Y la estimación de la historia es "5"

  # Caso LÍMITE (CA-16.1 y CA-16.2): todos los valores de la escala valen; fuera de ella, o vacío, no
  @CA-16.1 @CA-16.2
  Esquema del escenario: Valores de la escala
    Cuando estimo la historia en "<puntos>" story points
    Entonces la estimación da como resultado "<resultado>"

    Ejemplos:
      | puntos | resultado                                                                |
      | 0      | ok                                                                       |
      | 1      | ok                                                                       |
      | 2      | ok                                                                       |
      | 3      | ok                                                                       |
      | 5      | ok                                                                       |
      | 8      | ok                                                                       |
      | 13     | ok                                                                       |
      | 21     | ok                                                                       |
      | 4      | story_points: los story points tienen que ser 0, 1, 2, 3, 5, 8, 13 o 21 |
      | 22     | story_points: los story points tienen que ser 0, 1, 2, 3, 5, 8, 13 o 21 |
      | -1     | story_points: los story points tienen que ser 0, 1, 2, 3, 5, 8, 13 o 21 |
      | 3,5    | story_points: los story points tienen que ser 0, 1, 2, 3, 5, 8, 13 o 21 |
      | abc    | story_points: los story points tienen que ser 0, 1, 2, 3, 5, 8, 13 o 21 |
      |        | story_points: los story points tienen que ser 0, 1, 2, 3, 5, 8, 13 o 21 |

  # Caso LÍMITE (RN4): 0 es una estimación; "sin estimar" es no tener ninguna
  @CA-16.1
  Escenario: Cero cuenta como estimada
    Cuando estimo la historia en "0" story points
    Entonces la estimación de la historia es "0"

  # Caso de ERROR (CA-16.2 y RN7): un valor inválido no cambia la estimación que ya tenía
  @CA-16.2
  Escenario: Un valor fuera de la escala conserva la estimación anterior
    Dado que la historia ya está estimada en "5" story points
    Cuando estimo la historia en "4" story points
    Entonces la estimación da como resultado "story_points: los story points tienen que ser 0, 1, 2, 3, 5, 8, 13 o 21"
    Y la estimación de la historia es "5"

  # Caso LÍMITE (RN6): se estima hasta que la historia está Hecha
  @CA-16.1
  Esquema del escenario: Estimar según el estado de la historia
    Dado que la historia a estimar está "<estado>"
    Cuando estimo la historia en "8" story points
    Entonces la estimación da como resultado "<resultado>"

    Ejemplos:
      | estado      | resultado                                                              |
      | Pendiente   | ok                                                                     |
      | En sprint   | ok                                                                     |
      | En progreso | ok                                                                     |
      | Hecho       | la historia está Hecha: no se puede cambiar su estimación              |
