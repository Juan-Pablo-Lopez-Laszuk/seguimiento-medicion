# language: es
@HU-01
Característica: Crear proyecto
  Como Agile Enabler
  Quiero crear un proyecto con nombre, descripción, fecha de inicio y de finalización
  Para comenzar a gestionarlo

  # Caso NORMAL (CA-01.1 y CA-01.3)
  @CA-01.1 @CA-01.3
  Escenario: Crear un proyecto con datos válidos
    Cuando creo el proyecto "Software Metrics" con descripción "TPI de ICS" del "2026-10-05" al "2026-11-01"
    Entonces el proyecto "Software Metrics" queda creado
    Y su estado es "Planificado"

  # Caso ALTERNATIVO: la descripción es opcional y el nombre se guarda sin espacios sobrantes
  @CA-01.1
  Escenario: Crear un proyecto sin descripción y con espacios alrededor del nombre
    Cuando creo el proyecto "   Gestión Ñandú   " sin descripción del "2026-10-05" al "2026-11-01"
    Entonces el proyecto "Gestión Ñandú" queda creado

  # Caso LÍMITE: largo del nombre (se cuentan caracteres)
  @CA-01.1 @CA-01.4
  Esquema del escenario: Largo del nombre en los bordes
    Cuando creo un proyecto con un nombre de <largo> caracteres del "2026-10-05" al "2026-11-01"
    Entonces el resultado es "<resultado>"

    Ejemplos:
      | largo | resultado                                                |
      | 2     | nombre: el nombre debe tener entre 3 y 100 caracteres    |
      | 3     | creado                                                   |
      | 100   | creado                                                   |
      | 101   | nombre: el nombre debe tener entre 3 y 100 caracteres    |

  # Caso LÍMITE: fechas en el borde (CA-01.2)
  @CA-01.2 @CA-01.4
  Esquema del escenario: Fecha de finalización en el borde
    Cuando creo el proyecto "Borde de fechas" sin descripción del "<inicio>" al "<fin>"
    Entonces el resultado es "<resultado>"

    Ejemplos:
      | inicio     | fin        | resultado                                                              |
      | 2026-10-05 | 2026-10-05 | fecha_fin: la fecha de finalización debe ser posterior a la de inicio |
      | 2026-10-05 | 2026-10-06 | creado                                                                 |
      | 2026-10-05 | 2026-10-04 | fecha_fin: la fecha de finalización debe ser posterior a la de inicio |

  # Caso de ERROR: nombre repetido sin distinguir mayúsculas (CA-01.1)
  @CA-01.1 @CA-01.4
  Escenario: No se permite un nombre repetido
    Dado que existe el proyecto "Software Metrics"
    Cuando creo el proyecto "software metrics" sin descripción del "2026-10-05" al "2026-11-01"
    Entonces se informa el error "ya existe un proyecto con ese nombre" en el campo "nombre"
    Y hay 1 proyecto con el nombre "Software Metrics"

  # Caso de ERROR: varios datos inválidos se informan juntos (CA-01.4)
  @CA-01.4
  Escenario: Se indica cada campo que falló
    Cuando creo el proyecto "" sin descripción del "" al "2026-11-01"
    Entonces se informa el error "el nombre es obligatorio" en el campo "nombre"
    Y se informa el error "la fecha de inicio es obligatoria y debe ser válida" en el campo "fecha_inicio"
    Y no se crea ningún proyecto
