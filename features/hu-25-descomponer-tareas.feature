# language: es
@HU-25
Característica: Descomponer historia en tareas
  Como Product Builder
  Quiero dividir una historia en tareas
  Para organizar el trabajo

  Antecedentes:
    Dado el proyecto "Software Metrics" con la historia "Iniciar sesión con Google" y los integrantes "Carolina, Mariano"

  # Caso NORMAL (CA-25.1): la tarea tiene título, responsable y horas, y nace pendiente
  @CA-25.1
  Escenario: Agregar la primera tarea de la historia
    Cuando agrego la tarea "Pantalla de login" para "Carolina" con "3" horas
    Entonces la historia tiene la tarea "Pantalla de login" de "Carolina" con "3" horas, pendiente
    Y la historia suma "3" horas estimadas

  # Caso ALTERNATIVO (CA-25.1 y CA-23.2): varias tareas, en orden, y la historia suma sus horas
  @CA-25.1
  Escenario: Agregar varias tareas
    Dado que la historia tiene la tarea "Pantalla de login" de "Carolina" con "3" horas
    Cuando agrego la tarea "Guardar la sesión" para "Mariano" con "2,5" horas
    Entonces las tareas de la historia, en orden, son "Pantalla de login; Guardar la sesión"
    Y la historia suma "5,5" horas estimadas

  # Caso de ERROR (CA-25.1): cada dato es obligatorio y se valida
  @CA-25.1
  Esquema del escenario: Datos inválidos de una tarea
    Cuando agrego la tarea "<titulo>" para "<responsable>" con "<horas>" horas
    Entonces la tarea se rechaza con el error "<error>"
    Y la historia no tiene tareas

    Ejemplos:
      | titulo            | responsable | horas | error                                     |
      |                   | Carolina    | 3     | el título de la tarea es obligatorio      |
      | Pantalla de login |             | 3     | el responsable es obligatorio             |
      | Pantalla de login | Carolina    | 0     | las horas estimadas deben ser mayores a 0 |
      | Pantalla de login | Carolina    | abc   | las horas estimadas deben ser mayores a 0 |

  # Caso LÍMITE (CA-25.1): el título tiene hasta 120 caracteres
  @CA-25.1
  Esquema del escenario: Largo del título de la tarea
    Cuando agrego a la historia una tarea con un título de <largo> caracteres
    Entonces la tarea da como resultado "<resultado>"

    Ejemplos:
      | largo | resultado                                                  |
      | 120   | agregada                                                   |
      | 121   | el título de la tarea no puede superar los 120 caracteres  |

  # Caso de ERROR (CA-25.1): el responsable tiene que ser un integrante activo del proyecto
  @CA-25.1
  Escenario: Un integrante dado de baja no puede ser responsable
    Dado que "Mariano" está dado de baja
    Cuando agrego la tarea "Guardar la sesión" para "Mariano" con "2" horas
    Entonces la tarea se rechaza con el error "el responsable tiene que ser un integrante activo del proyecto"

  # Caso ALTERNATIVO (CA-25.2): cada historia tiene sus propias tareas
  @CA-25.2
  Escenario: Cada historia tiene sus propias tareas
    Dado la historia "Ver el tablero" con la tarea "Armar la tabla" de "Mariano" con "4" horas
    Cuando agrego la tarea "Pantalla de login" para "Carolina" con "3" horas
    Entonces las tareas de la historia, en orden, son "Pantalla de login"
    Y la historia suma "3" horas estimadas
    Y la historia "Ver el tablero" suma "4" horas estimadas

  # Caso de ERROR (CA-25.2): una tarea no se modifica desde otra historia
  @CA-25.2
  Escenario: Una tarea no se cambia desde otra historia
    Dado la historia "Ver el tablero" con la tarea "Armar la tabla" de "Mariano" con "4" horas
    Cuando cambio a "6" las horas de la tarea "Armar la tabla" desde la historia "Iniciar sesión con Google"
    Entonces no se encuentra la tarea
    Y la historia "Ver el tablero" suma "4" horas estimadas

  # Caso NORMAL (CA-25.3): mientras está pendiente, la estimación se puede cambiar
  @CA-25.3
  Escenario: Cambiar la estimación de una tarea pendiente
    Dado que la historia tiene la tarea "Pantalla de login" de "Carolina" con "3" horas
    Cuando cambio a "4" las horas de la tarea "Pantalla de login"
    Entonces la historia suma "4" horas estimadas

  # Caso de ERROR (CA-25.3): una tarea completada no cambia su estimación
  @CA-25.3
  Escenario: Una tarea completada no cambia su estimación
    Dado que la historia tiene la tarea "Pantalla de login" de "Carolina" con "3" horas
    Y que la tarea "Pantalla de login" está completada
    Cuando cambio a "5" las horas de la tarea "Pantalla de login"
    Entonces la tarea se rechaza con el error "la tarea está completada: no se puede cambiar su estimación"
    Y la historia suma "3" horas estimadas

  # Caso LÍMITE (CA-25.3): completar no cambia las horas, y volver a pendiente permite corregirlas
  @CA-25.3
  Escenario: Completar una tarea y volver a pendiente para corregir las horas
    Dado que la historia tiene la tarea "Pantalla de login" de "Carolina" con "3" horas
    Cuando completo la tarea "Pantalla de login"
    Entonces la tarea "Pantalla de login" queda completada
    Y la historia suma "3" horas estimadas
    Cuando vuelvo a pendiente la tarea "Pantalla de login"
    Y cambio a "2" las horas de la tarea "Pantalla de login"
    Entonces la historia suma "2" horas estimadas
