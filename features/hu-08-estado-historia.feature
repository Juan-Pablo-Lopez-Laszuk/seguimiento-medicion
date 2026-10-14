# language: es
@HU-08
Característica: Cambiar el estado de una historia
  Como Product Builder
  Quiero cambiar el estado de una historia
  Para reflejar su avance

  Antecedentes:
    Dado el proyecto "Software Metrics" con la historia "Iniciar sesión con Google" para cambiar su estado

  # Caso NORMAL (CA-08.1): el paso que se hace desde la página de la historia
  @CA-08.1
  Escenario: Una historia En sprint pasa a En progreso
    Dado que la historia está "En sprint"
    Cuando la paso a "En progreso"
    Entonces el estado de la historia es "En progreso"

  # Caso ALTERNATIVO (CA-08.1 y CA-08.2): el camino completo, con todos los criterios cumplidos
  @CA-08.1 @CA-08.2
  Escenario: Una historia recorre todos los estados hasta Hecho
    Dado que la historia está "Pendiente" y tiene 2 criterios, 2 de ellos cumplidos
    Cuando la paso a "En sprint"
    Y la paso a "En progreso"
    Y la paso a "Hecho"
    Entonces el estado de la historia es "Hecho"

  # Caso LÍMITE (CA-08.1 y CA-08.3): cada par de estados, los que valen y los que no, con su mensaje
  @CA-08.1 @CA-08.3
  Esquema del escenario: Qué transiciones valen
    Dado que la historia está "<desde>" y tiene 2 criterios, 2 de ellos cumplidos
    Cuando la paso a "<hacia>"
    Entonces el cambio da como resultado "<resultado>"

    Ejemplos:
      | desde       | hacia       | resultado                                                                                       |
      | Pendiente   | En sprint   | ok                                                                                              |
      | En sprint   | En progreso | ok                                                                                              |
      | En progreso | Hecho       | ok                                                                                              |
      | En sprint   | Pendiente   | ok                                                                                              |
      | En progreso | Pendiente   | ok                                                                                              |
      | Pendiente   | En progreso | no se puede pasar de Pendiente a En progreso: desde Pendiente solo se puede pasar a En sprint   |
      | Pendiente   | Hecho       | no se puede pasar de Pendiente a Hecho: desde Pendiente solo se puede pasar a En sprint         |
      | En sprint   | Hecho       | no se puede pasar de En sprint a Hecho: desde En sprint solo se puede pasar a En progreso o Pendiente |
      | En progreso | En sprint   | no se puede pasar de En progreso a En sprint: desde En progreso solo se puede pasar a Hecho o Pendiente |
      | Hecho       | En progreso | no se puede pasar de Hecho a En progreso: una historia Hecha no cambia de estado                |
      | Hecho       | Pendiente   | no se puede pasar de Hecho a Pendiente: una historia Hecha no cambia de estado                  |

  # Caso LÍMITE (CA-08.2): a Hecho solo con criterios y todos cumplidos; sin criterios no
  @CA-08.2
  Esquema del escenario: Para pasar a Hecho hacen falta todos los criterios cumplidos
    Dado que la historia está "En progreso" y tiene <total> criterios, <cumplidos> de ellos cumplidos
    Cuando la paso a "Hecho"
    Entonces el cambio da como resultado "<resultado>"

    Ejemplos:
      | total | cumplidos | resultado                                                                                  |
      | 0     | 0         | no se puede pasar a Hecho: la historia no tiene criterios de aceptación                    |
      | 3     | 0         | no se puede pasar a Hecho: la historia tiene 0 de 3 criterios cumplidos y hacen falta todos |
      | 3     | 2         | no se puede pasar a Hecho: la historia tiene 2 de 3 criterios cumplidos y hacen falta todos |
      | 1     | 1         | ok                                                                                         |
      | 3     | 3         | ok                                                                                         |

  # Caso de ERROR (CA-08.2 y CA-08.3): si el cambio no vale, la historia queda como estaba (RN7)
  @CA-08.2 @CA-08.3
  Escenario: Un cambio inválido deja la historia como estaba
    Dado que la historia está "En progreso" y tiene 3 criterios, 2 de ellos cumplidos
    Cuando la paso a "Hecho"
    Entonces el cambio da como resultado "no se puede pasar a Hecho: la historia tiene 2 de 3 criterios cumplidos y hacen falta todos"
    Y el estado de la historia es "En progreso"

  # Caso de ERROR (CA-08.3): pedir el estado en el que ya está (RN5)
  @CA-08.3
  Escenario: Pasar a un estado en el que ya está
    Dado que la historia está "En progreso"
    Cuando la paso a "En progreso"
    Entonces el cambio da como resultado "la historia ya está En progreso"

  # Caso de ERROR (CA-08.3): un estado que no existe (RN1)
  @CA-08.3
  Escenario: Pasar a un estado que no existe
    Dado que la historia está "En sprint"
    Cuando la paso a "Terminada"
    Entonces el cambio da como resultado "el estado Terminada no existe: los estados son Pendiente, En sprint, En progreso y Hecho"
    Y el estado de la historia es "En sprint"
