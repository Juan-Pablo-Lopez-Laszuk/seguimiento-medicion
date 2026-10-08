# language: es
@HU-11
Característica: Crear sprint con Sprint Goal
  Como Agile Enabler
  Quiero crear un sprint con fechas y Sprint Goal
  Para planificar la iteración

  Antecedentes:
    Dado el proyecto "Software Metrics" que va del "2026-10-05" al "2026-11-01"

  # Caso NORMAL (CA-11.1 y CA-11.3)
  @CA-11.1 @CA-11.3
  Escenario: Crear el primer sprint del proyecto
    Cuando creo un sprint con el goal "Primer MVP" del "2026-10-05" al "2026-10-11"
    Entonces se crea el Sprint 1 con el goal "Primer MVP"
    Y el sprint queda en estado "Planificado"

  # Caso ALTERNATIVO: el número sigue al último y puede haber días libres entre sprints (CA-11.3)
  @CA-11.3
  Escenario: El siguiente sprint lleva el número que sigue
    Dado que el proyecto ya tiene un sprint del "2026-10-05" al "2026-10-11"
    Cuando creo un sprint con el goal "Interfaz" del "2026-10-14" al "2026-10-20"
    Entonces se crea el Sprint 2 con el goal "Interfaz"

  # Caso ALTERNATIVO: cada proyecto tiene su propia numeración (CA-11.3)
  @CA-11.3
  Escenario: Cada proyecto numera sus propios sprints
    Dado que el proyecto ya tiene un sprint del "2026-10-05" al "2026-10-11"
    Y el proyecto "Otro proyecto" que va del "2026-10-01" al "2026-12-31"
    Cuando creo en "Otro proyecto" un sprint con el goal "Arranque" del "2026-10-01" al "2026-10-07"
    Entonces se crea el Sprint 1 con el goal "Arranque"

  # Caso LÍMITE: los bordes del proyecto valen; un día afuera, no (CA-11.2)
  @CA-11.2
  Esquema del escenario: Fechas en el borde del proyecto
    Cuando creo un sprint con el goal "Borde" del "<inicio>" al "<fin>"
    Entonces el sprint da como resultado "<resultado>"

    Ejemplos:
      | inicio     | fin        | resultado                                                                                 |
      | 2026-10-05 | 2026-11-01 | creado                                                                                    |
      | 2026-10-04 | 2026-10-11 | fecha_inicio: la fecha tiene que estar dentro del proyecto (del 05/10/2026 al 01/11/2026) |
      | 2026-10-26 | 2026-11-02 | fecha_fin: la fecha tiene que estar dentro del proyecto (del 05/10/2026 al 01/11/2026)    |
      | 2026-10-05 | 2026-10-05 | fecha_fin: la fecha de finalización debe ser posterior a la de inicio                     |

  # Caso de ERROR: se superpone con el último sprint o queda antes (CA-11.2)
  @CA-11.2
  Esquema del escenario: No se superpone con el último sprint
    Dado que el proyecto ya tiene un sprint del "2026-10-05" al "2026-10-11"
    Cuando creo un sprint con el goal "Superpuesto" del "<inicio>" al "<fin>"
    Entonces el sprint da como resultado "<resultado>"

    Ejemplos:
      | inicio     | fin        | resultado                                                                                    |
      | 2026-10-11 | 2026-10-18 | fecha_inicio: el sprint tiene que empezar después del último (el Sprint 1 termina el 11/10/2026) |
      | 2026-10-08 | 2026-10-10 | fecha_inicio: el sprint tiene que empezar después del último (el Sprint 1 termina el 11/10/2026) |
      | 2026-10-12 | 2026-10-18 | creado                                                                                       |

  # Caso de ERROR: Sprint Goal vacío; todos los errores se informan juntos (CA-11.1)
  @CA-11.1
  Escenario: El Sprint Goal es obligatorio
    Cuando creo un sprint con el goal "   " del "" al "2026-10-11"
    Entonces el sprint da como resultado "objetivo: el Sprint Goal es obligatorio"
    Y el sprint da como resultado "fecha_inicio: la fecha de inicio es obligatoria y debe ser válida"
    Y el proyecto no tiene sprints
