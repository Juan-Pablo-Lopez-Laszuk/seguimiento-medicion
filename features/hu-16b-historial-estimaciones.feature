# language: es
@HU-16b
Característica: Historial de estimaciones
  Como equipo
  Queremos que al re-estimar una historia se conserve el valor anterior
  Para ver cómo cambió su estimación

  Antecedentes:
    Dado el proyecto "Software Metrics" con las historias "Iniciar sesión con Google" y "Cerrar sesión" para el historial

  # Caso NORMAL (CA-16.3): re-estimar anota el valor anterior
  @CA-16.3
  Escenario: Re-estimar conserva el valor anterior
    Dado que "Iniciar sesión con Google" está estimada en 5 story points
    Cuando cambio la estimación de "Iniciar sesión con Google" a "8"
    Entonces el historial de "Iniciar sesión con Google" es:
      | anterior | nuevo |
      | 5        | 8     |
    Y la estimación de "Iniciar sesión con Google" es "8"

  # Caso ALTERNATIVO (CA-16.3): varios cambios, el más reciente primero
  @CA-16.3
  Escenario: Varias re-estimaciones, la más reciente primero
    Dado que "Iniciar sesión con Google" está estimada en 5 story points
    Cuando cambio la estimación de "Iniciar sesión con Google" a "8"
    Y cambio la estimación de "Iniciar sesión con Google" a "3"
    Entonces el historial de "Iniciar sesión con Google" es:
      | anterior | nuevo |
      | 8        | 3     |
      | 5        | 8     |

  # Caso LÍMITE (RN2): la primera estimación no tiene valor anterior
  @CA-16.3
  Escenario: La primera estimación no tiene valor anterior
    Cuando cambio la estimación de "Iniciar sesión con Google" a "5"
    Entonces el historial de "Iniciar sesión con Google" está vacío
    Y la estimación de "Iniciar sesión con Google" es "5"

  # Caso LÍMITE (RN3): pedir el mismo valor no cambia nada
  @CA-16.3
  Escenario: Pedir el mismo valor no anota nada
    Dado que "Iniciar sesión con Google" está estimada en 5 story points
    Cuando cambio la estimación de "Iniciar sesión con Google" a "5"
    Entonces el historial de "Iniciar sesión con Google" está vacío

  # Caso LÍMITE (RN4): 0 es una estimación, "sin estimar" no
  @CA-16.3
  Esquema del escenario: Cero cuenta como estimación
    Dado que "Iniciar sesión con Google" está estimada en <antes> story points
    Cuando cambio la estimación de "Iniciar sesión con Google" a "<despues>"
    Entonces el historial de "Iniciar sesión con Google" es:
      | anterior | nuevo      |
      | <antes>  | <despues>  |

    Ejemplos:
      | antes | despues |
      | 0     | 5       |
      | 5     | 0       |

  # Caso de ERROR (RN5): un valor inválido no anota nada y conserva la estimación
  @CA-16.3
  Escenario: Un valor inválido no anota nada
    Dado que "Iniciar sesión con Google" está estimada en 5 story points
    Cuando cambio la estimación de "Iniciar sesión con Google" a "4"
    Entonces el cambio da como resultado "story_points: los story points tienen que ser 0, 1, 2, 3, 5, 8, 13 o 21"
    Y el historial de "Iniciar sesión con Google" está vacío
    Y la estimación de "Iniciar sesión con Google" es "5"

  # Caso de ERROR (RN5): una historia Hecha no se re-estima, y su historial se conserva
  @CA-16.3
  Escenario: Una historia Hecha no se re-estima
    Dado que "Iniciar sesión con Google" está estimada en 5 story points
    Y que "Iniciar sesión con Google" se re-estimó a "8"
    Y que "Iniciar sesión con Google" está Hecha
    Cuando cambio la estimación de "Iniciar sesión con Google" a "13"
    Entonces el cambio da como resultado "la historia está Hecha: no se puede cambiar su estimación"
    Y el historial de "Iniciar sesión con Google" es:
      | anterior | nuevo |
      | 5        | 8     |

  # Caso LÍMITE (RN7): cada historia tiene su historial
  @CA-16.3
  Escenario: El historial es de cada historia
    Dado que "Iniciar sesión con Google" está estimada en 5 story points
    Y que "Cerrar sesión" está estimada en 2 story points
    Cuando cambio la estimación de "Iniciar sesión con Google" a "8"
    Entonces el historial de "Iniciar sesión con Google" es:
      | anterior | nuevo |
      | 5        | 8     |
    Y el historial de "Cerrar sesión" está vacío

  # Caso LÍMITE (RN9): al eliminar la historia se elimina su historial
  @RN9
  Escenario: Eliminar una historia con historial
    Dado que "Iniciar sesión con Google" está estimada en 5 story points
    Y que "Iniciar sesión con Google" se re-estimó a "8"
    Cuando elimino la historia "Iniciar sesión con Google"
    Entonces la historia "Iniciar sesión con Google" ya no existe
    Y no queda ningún historial de la historia eliminada
