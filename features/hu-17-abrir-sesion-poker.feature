# language: es
@HU-17
Característica: Abrir sesión de Planning Poker
  Como Agile Enabler
  Quiero abrir una sesión de Planning Poker para una historia
  Para que el equipo la estime en conjunto

  Antecedentes:
    Dado el proyecto "Software Metrics" con estos integrantes para Planning Poker:
      | nombre   | rol            | activo |
      | Juan     | AgileEnabler   | sí     |
      | Carolina | ProductBuilder | sí     |
      | Mariano  | ProductBuilder | sí     |
      | Lucía    | ProductBuilder | no     |
    Y estas historias para estimar:
      | título   | estado      |
      | Login    | Pendiente   |
      | Registro | En sprint   |
      | Perfil   | En progreso |
      | Reportes | Hecho       |

  # Caso NORMAL (CA-17.1, CA-17.3, RN3): el Agile Enabler abre la sesión y participan los integrantes activos
  @CA-17.1 @CA-17.3 @RN3
  Escenario: El Agile Enabler abre una sesión
    Cuando "Juan" abre una sesión de Planning Poker para "Login"
    Entonces la historia "Login" tiene una sesión de Planning Poker abierta en la ronda 1
    Y participan estos integrantes:
      | nombre   |
      | Juan     |
      | Carolina |
      | Mariano  |

  # Caso ALTERNATIVO (CA-17.1, RN1): se abre para cualquier historia que no está Hecha
  @CA-17.1 @RN1
  Esquema del escenario: Historias que se pueden estimar
    Cuando "Juan" abre una sesión de Planning Poker para "<historia>"
    Entonces la historia "<historia>" tiene una sesión de Planning Poker abierta en la ronda 1

    Ejemplos:
      | historia |
      | Login    |
      | Registro |
      | Perfil   |

  # Caso de ERROR (CA-17.1): una historia Hecha no se estima
  @CA-17.1
  Escenario: No se abre una sesión para una historia Hecha
    Cuando "Juan" abre una sesión de Planning Poker para "Reportes"
    Entonces abrir la sesión da el error "la historia está Hecha: ya no se estima"
    Y la historia "Reportes" no tiene una sesión de Planning Poker abierta

  # Caso de ERROR (CA-17.2, RN2): una sola sesión abierta por historia
  @CA-17.2 @RN2
  Escenario: No se abre una segunda sesión para la misma historia
    Dado que "Juan" abrió una sesión de Planning Poker para "Login"
    Cuando "Juan" abre una sesión de Planning Poker para "Login"
    Entonces abrir la sesión da el error "la historia ya tiene una sesión de Planning Poker abierta"
    Y la historia "Login" tiene una sola sesión de Planning Poker abierta

  # Caso de ERROR (RN4): solo el Agile Enabler activo abre la sesión
  @RN4
  Escenario: Un Product Builder no abre una sesión
    Cuando "Carolina" abre una sesión de Planning Poker para "Login"
    Entonces abrir la sesión da el error "solo el Agile Enabler del proyecto abre una sesión de Planning Poker"
    Y la historia "Login" no tiene una sesión de Planning Poker abierta

  # Caso LÍMITE (CA-17.3, RN5): los participantes son los activos de cada momento
  @CA-17.3 @RN5
  Escenario: Un integrante que se da de baja deja de participar
    Dado que "Juan" abrió una sesión de Planning Poker para "Login"
    Cuando se da de baja a "Mariano"
    Entonces participan estos integrantes:
      | nombre   |
      | Juan     |
      | Carolina |

  # Caso LÍMITE (RN4): sin un Agile Enabler activo nadie abre la sesión
  @RN4
  Escenario: Un Agile Enabler dado de baja no abre una sesión
    Dado que se da de baja a "Juan"
    Cuando "Juan" abre una sesión de Planning Poker para "Login"
    Entonces abrir la sesión da el error "solo el Agile Enabler del proyecto abre una sesión de Planning Poker"
