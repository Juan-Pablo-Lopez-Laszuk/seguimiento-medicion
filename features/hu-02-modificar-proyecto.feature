# language: es
@HU-02
Característica: Modificar proyecto
  Como Agile Enabler
  Quiero modificar los datos de un proyecto
  Para mantener su información actualizada

  Antecedentes:
    Dado el proyecto a modificar "Software Metrics" del "2026-10-05" al "2026-11-01"
    Y que ese proyecto tiene el Sprint 1 del "2026-10-05" al "2026-10-11" y el Sprint 2 del "2026-10-12" al "2026-10-18"
    Y que también existe el proyecto "Otro proyecto"
    Y que hoy es el "2026-10-13" para modificar

  # Caso NORMAL (CA-02.1)
  @CA-02.1
  Escenario: Modificar la descripción
    Cuando modifico el proyecto con nombre "Software Metrics", descripción "TPI de ICS 2026" del "2026-10-05" al "2026-11-01"
    Entonces el proyecto queda con la descripción "TPI de ICS 2026"

  # Caso ALTERNATIVO: el propio nombre con otras mayúsculas no es un nombre repetido (CA-02.1)
  @CA-02.1
  Escenario: Cambiar las mayúsculas del propio nombre
    Cuando modifico el proyecto con nombre "software METRICS", descripción "" del "2026-10-05" al "2026-11-01"
    Entonces el proyecto queda con el nombre "software METRICS"

  # Caso de ERROR: mismas validaciones que al crear, todas juntas (CA-02.1)
  @CA-02.1
  Escenario: Se informan los errores de la creación
    Cuando modifico el proyecto con nombre "OTRO PROYECTO", descripción "" del "2026-10-05" al "2026-10-05"
    Entonces la modificación da como resultado "nombre: ya existe un proyecto con ese nombre"
    Y la modificación da como resultado "fecha_fin: la fecha de finalización debe ser posterior a la de inicio"
    Y el proyecto sigue con el nombre "Software Metrics"

  # Caso LÍMITE: las fechas del proyecto contra sus sprints (CA-02.2)
  @CA-02.2
  Esquema del escenario: Fechas contra los sprints
    Cuando modifico el proyecto con nombre "Software Metrics", descripción "" del "<inicio>" al "<fin>"
    Entonces la modificación da como resultado "<resultado>"

    Ejemplos:
      | inicio     | fin        | resultado                                                                                                               |
      | 2026-10-05 | 2026-10-18 | modificado                                                                                                              |
      | 2026-10-05 | 2026-10-17 | fecha_fin: la fecha de finalización no puede ser anterior al fin del último sprint (el Sprint 2 termina el 18/10/2026)  |
      | 2026-10-01 | 2026-11-01 | modificado                                                                                                              |
      | 2026-10-06 | 2026-11-01 | fecha_inicio: la fecha de inicio no puede ser posterior al inicio del primer sprint (el Sprint 1 empieza el 05/10/2026) |

  # Caso de ERROR: un proyecto Finalizado no se modifica (CA-02.3)
  @CA-02.3
  Escenario: Un proyecto Finalizado no se modifica
    Dado que hoy es el "2026-11-02" para modificar
    Cuando modifico el proyecto con nombre "Software Metrics", descripción "Cambio tardío" del "2026-10-05" al "2026-11-01"
    Entonces la modificación da el error "un proyecto Finalizado no se puede modificar"
    Y el proyecto sigue con la descripción ""
