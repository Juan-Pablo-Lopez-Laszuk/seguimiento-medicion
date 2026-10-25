# language: es
@HU-19
Característica: Revelar estimaciones
  Como equipo
  Quiero ver todas las estimaciones al finalizar la votación
  Para discutirlas

  Antecedentes:
    Dado el equipo de Planning Poker con "Juan Pablo" de Agile Enabler y "Carolina" y "Mariano" de Product Builder
    Y una sesión de Planning Poker abierta para "HU-5 Login" con la ronda 1

  # Caso NORMAL (CA-19.1 y CA-19.2): votaron todos y se revela sola
  @CA-19.1 @CA-19.2
  Escenario: La ronda se revela cuando votaron todos
    Dado que en la ronda votaron "Juan Pablo" 5, "Carolina" 8 y "Mariano" 5
    Entonces la ronda está revelada
    Y se ven estos votos:
      | integrante | voto |
      | Carolina   | 8    |
      | Juan Pablo | 5    |
      | Mariano    | 5    |

  # Caso ALTERNATIVO (CA-19.1): el Agile Enabler cierra antes
  @CA-19.1 @CA-19.2
  Escenario: El Agile Enabler cierra la ronda antes de que voten todos
    Dado que en la ronda votaron "Carolina" 3 y "Mariano" 5
    Cuando "Juan Pablo" cierra la ronda
    Entonces la ronda está revelada
    Y se ven estos votos:
      | integrante | voto     |
      | Carolina   | 3        |
      | Juan Pablo | no votó  |
      | Mariano    | 5        |

  # Caso LÍMITE (CA-19.1): falta un voto y nadie cierra
  @CA-19.1
  Escenario: Falta un voto y la ronda sigue oculta
    Dado que en la ronda votaron "Juan Pablo" 5 y "Carolina" 8
    Entonces la ronda no está revelada
    Y no se ve ningún voto

  # Caso de ERROR (CA-19.3): una ronda sin votos no se revela
  @CA-19.3
  Escenario: Una ronda sin votos no se puede revelar
    Cuando "Juan Pablo" cierra la ronda
    Entonces revelar da el error "la ronda no tiene votos: no se puede revelar"

  # Caso de ERROR (RN4): solo el Agile Enabler cierra la ronda
  @CA-19.1
  Escenario: Un Product Builder no puede cerrar la ronda
    Dado que en la ronda votaron "Carolina" 3
    Cuando "Mariano" cierra la ronda
    Entonces revelar da el error "solo el Agile Enabler del proyecto abre una sesión de Planning Poker"
