# language: es
@HU-18
Característica: Votar de forma oculta
  Como integrante
  Quiero emitir mi estimación sin que los demás la vean
  Para evitar el efecto ancla

  Antecedentes:
    Dado el proyecto "Software Metrics" con una sesión de Planning Poker abierta para "Login" y estos integrantes:
      | nombre   | rol            | activo |
      | Juan     | AgileEnabler   | sí     |
      | Carolina | ProductBuilder | sí     |
      | Mariano  | ProductBuilder | sí     |
      | Lucía    | ProductBuilder | no     |

  # Caso NORMAL (CA-18.1, CA-18.2): se ve quién votó
  @CA-18.1 @CA-18.2
  Escenario: Un integrante vota
    Cuando "Carolina" vota "5"
    Entonces en la ronda votaron 1 de 3
    Y votaron estos integrantes:
      | nombre   |
      | Carolina |

  # Caso ALTERNATIVO (CA-18.1, RN1): votar de nuevo reemplaza el voto
  @CA-18.1 @RN1
  Escenario: Cambiar el voto antes de revelar
    Dado que "Carolina" votó "5"
    Cuando "Carolina" vota "8"
    Entonces en la ronda votaron 1 de 3

  # Caso ALTERNATIVO (RN2): el Agile Enabler también vota
  @RN2
  Escenario: Votan todos los participantes
    Cuando "Juan" vota "3"
    Y "Carolina" vota "5"
    Y "Mariano" vota "8"
    Entonces en la ronda votaron 3 de 3

  # Caso de ERROR (CA-18.2, CA-18.3, RN5): los valores no salen del servidor antes de revelar
  @CA-18.2 @CA-18.3 @RN5
  Escenario: Nadie ve los valores antes de revelar
    Dado que "Carolina" votó "13"
    Y que "Mariano" votó "21"
    Entonces la lista de quién votó no muestra ningún valor
    Y la vista de la sesión no tiene ningún valor de voto

  # Caso LÍMITE (CA-18.4, RN4): los extremos de la escala valen
  @CA-18.4 @RN4
  Esquema del escenario: Valores de la escala
    Cuando "Carolina" vota "<valor>"
    Entonces en la ronda votaron 1 de 3

    Ejemplos:
      | valor |
      | 0     |
      | 21    |

  # Caso de ERROR (CA-18.4, RN4): un valor fuera de la escala no se guarda
  @CA-18.4 @RN4
  Esquema del escenario: Voto fuera de la escala
    Cuando "Carolina" vota "<valor>"
    Entonces votar da el error "el voto tiene que ser un valor de la escala: 0, 1, 2, 3, 5, 8, 13 o 21"
    Y en la ronda votaron 0 de 3

    Ejemplos:
      | valor |
      | 4     |
      | -1    |
      | abc   |
      |       |

  # Caso de ERROR (RN2): una integrante dada de baja no vota
  @RN2
  Escenario: Una integrante dada de baja no vota
    Cuando "Lucía" vota "5"
    Entonces votar da el error "solo votan los integrantes activos del proyecto"
    Y en la ronda votaron 0 de 3

  # Caso LÍMITE (RN2): el voto de quien se da de baja deja de contar
  @RN2
  Escenario: El voto de un integrante que se da de baja no cuenta
    Dado que "Mariano" votó "8"
    Cuando "Mariano" queda dado de baja
    Entonces en la ronda votaron 0 de 2

  # Caso de ERROR (RN3): sin sesión abierta no se vota
  @RN3
  Escenario: No se vota en una historia sin sesión abierta
    Cuando "Carolina" vota "5" en la historia "Perfil", que no tiene sesión
    Entonces votar da el error "esta historia no tiene una sesión de Planning Poker abierta"
