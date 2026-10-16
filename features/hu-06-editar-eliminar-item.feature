# language: es
@HU-06
Característica: Editar y eliminar ítem del Product Backlog
  Como Product Builder
  Quiero editar o eliminar historias
  Para mantener el backlog ordenado

  Antecedentes:
    Dado el proyecto "Software Metrics" con la historia "Iniciar sesión con Google" para editar o eliminar

  # Caso NORMAL (CA-06.1): editar cambia los datos y no el identificador
  @CA-06.1
  Escenario: Editar los datos no cambia el identificador
    Cuando edito la historia con título "Iniciar sesión con Google o GitHub", descripción "Con OAuth" y prioridad "Baja"
    Entonces la edición da como resultado "ok"
    Y la historia es HU-1 con título "Iniciar sesión con Google o GitHub", descripción "Con OAuth" y prioridad "Baja"

  # Caso ALTERNATIVO (RN2): editar no toca el estado, los story points, los criterios ni las tareas
  @CA-06.1
  Escenario: Editar no cambia el estado, los story points, los criterios ni las tareas
    Dado que la historia a editar tiene "5" story points, 2 criterios, 1 tarea y está "En progreso"
    Cuando edito la historia con título "Otro título", descripción "" y prioridad "Media"
    Entonces la historia conserva "5" story points, 2 criterios, 1 tarea y el estado "En progreso"

  # Caso LÍMITE (RN4): guardar sin haber cambiado nada vale
  @CA-06.1
  Escenario: Guardar sin cambios
    Cuando edito la historia con título "Iniciar sesión con Google", descripción "" y prioridad "Alta"
    Entonces la edición da como resultado "ok"
    Y la historia es HU-1 con título "Iniciar sesión con Google", descripción "" y prioridad "Alta"

  # Caso LÍMITE (RN1): el título y la descripción se guardan sin los espacios de los extremos
  @CA-06.1
  Escenario: Los espacios de los extremos se quitan
    Cuando edito la historia con título "  Login con Google  ", descripción "  Con OAuth  " y prioridad "Alta"
    Entonces la historia es HU-1 con título "Login con Google", descripción "Con OAuth" y prioridad "Alta"

  # Caso de ERROR (RN1): cada dato inválido da su error y la historia no cambia
  @CA-06.1
  Esquema del escenario: Datos inválidos
    Cuando edito la historia con título "<titulo>", descripción "" y prioridad "<prioridad>"
    Entonces la edición da como resultado "<resultado>"
    Y la historia es HU-1 con título "Iniciar sesión con Google", descripción "" y prioridad "Alta"

    Ejemplos:
      | titulo | prioridad | resultado                                              |
      |        | Alta      | titulo: el título es obligatorio                       |
      |        | Media     | titulo: el título es obligatorio                       |
      | Login  | Urgente   | prioridad: la prioridad tiene que ser Alta, Media o Baja |
      | Login  |           | prioridad: la prioridad tiene que ser Alta, Media o Baja |

  # Caso de ERROR (RN1): si hay varios datos inválidos se informan todos juntos
  @CA-06.1
  Escenario: Varios errores a la vez
    Cuando edito la historia con título "", descripción "" y prioridad "Urgente"
    Entonces la edición da como resultado "titulo: el título es obligatorio"
    Y la edición da como resultado "prioridad: la prioridad tiene que ser Alta, Media o Baja"

  # Caso LÍMITE (RN1): el largo máximo del título y de la descripción
  @CA-06.1
  Esquema del escenario: Largo del título y de la descripción
    Cuando edito la historia con <campo> de <caracteres> caracteres
    Entonces la edición da como resultado "<resultado>"

    Ejemplos:
      | campo       | caracteres | resultado                                            |
      | título      | 120        | ok                                                   |
      | título      | 121        | titulo: el título no puede superar los 120 caracteres |
      | descripción | 2000       | ok                                                   |
      | descripción | 2001       | descripcion: la descripción no puede superar los 2000 caracteres |

  # Caso LÍMITE (RN3): se puede editar en cualquier estado, también una historia Hecha
  @CA-06.1
  Esquema del escenario: Editar según el estado de la historia
    Dado que la historia a editar está "<estado>"
    Cuando edito la historia con título "Título corregido", descripción "" y prioridad "Alta"
    Entonces la edición da como resultado "ok"

    Ejemplos:
      | estado      |
      | Pendiente   |
      | En sprint   |
      | En progreso |
      | Hecho       |

  # Caso NORMAL (CA-06.3, RN8): eliminar primero pide confirmación y no borra nada
  @CA-06.3
  Escenario: Eliminar pide confirmación
    Dado que la historia a editar tiene "3" story points, 2 criterios, 1 tarea y está "Pendiente"
    Cuando pido eliminar la historia
    Entonces el sistema pide confirmar la eliminación de 2 criterios y 1 tarea
    Y la historia sigue en el backlog

  # Caso NORMAL (CA-06.3, RN9): al confirmar se elimina la historia con sus criterios y sus tareas
  @CA-06.3
  Escenario: Confirmar elimina la historia con sus criterios y sus tareas
    Dado que la historia a editar tiene "3" story points, 2 criterios, 1 tarea y está "Pendiente"
    Cuando confirmo la eliminación de la historia
    Entonces la eliminación da como resultado "ok"
    Y la historia ya no está en el backlog
    Y no quedan criterios ni tareas de la historia

  # Caso LÍMITE (RN7): se puede eliminar mientras no esté Hecha ni sea de un sprint cerrado
  @CA-06.2
  Esquema del escenario: Eliminar según el estado de la historia
    Dado que la historia a editar está "<estado>"
    Cuando confirmo la eliminación de la historia
    Entonces la eliminación da como resultado "<resultado>"
    Y la historia <situacion>

    Ejemplos:
      | estado      | resultado                                         | situacion                   |
      | Pendiente   | ok                                                | ya no está en el backlog    |
      | En sprint   | ok                                                | ya no está en el backlog    |
      | En progreso | ok                                                | ya no está en el backlog    |
      | Hecho       | la historia está Hecha: no se puede eliminar      | sigue en el backlog         |

  # Caso de ERROR (CA-06.2): una historia de un sprint cerrado no se elimina, aunque no esté Hecha
  @CA-06.2
  Escenario: Una historia de un sprint cerrado no se elimina
    Dado que la historia a editar es de un sprint cerrado
    Cuando confirmo la eliminación de la historia
    Entonces la eliminación da como resultado "la historia es de un sprint cerrado: no se puede eliminar"
    Y la historia sigue en el backlog

  # Caso de ERROR (CA-06.2): una historia que no se puede eliminar lo dice ya en la pantalla de confirmación
  @CA-06.2
  Escenario: La confirmación explica por qué no se puede eliminar
    Dado que la historia a editar está "Hecho"
    Cuando pido eliminar la historia
    Entonces el sistema no deja eliminarla: "la historia está Hecha: no se puede eliminar"

  # Caso LÍMITE (RN11): las demás historias conservan su número y la última libera el suyo
  @CA-06.1
  Escenario: Eliminar una historia del medio no renumera las demás
    Dado que el proyecto tiene además las historias "Registro con email" y "Recuperar contraseña"
    Cuando confirmo la eliminación de la historia "Registro con email"
    Entonces las historias del proyecto son "HU-1, HU-3"
    Y al crear una historia nueva es la HU-4

  @CA-06.1
  Escenario: Eliminar la última historia libera su número
    Dado que el proyecto tiene además las historias "Registro con email" y "Recuperar contraseña"
    Cuando confirmo la eliminación de la historia "Recuperar contraseña"
    Entonces las historias del proyecto son "HU-1, HU-2"
    Y al crear una historia nueva es la HU-3
