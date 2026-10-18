# language: es
@HU-24
Característica: Registrar esfuerzo realizado
  Como integrante
  Quiero cargar fecha, actividad y horas trabajadas
  Para medir el esfuerzo real

  Antecedentes:
    Dado el proyecto del 05/10/2026 al 01/11/2026 con la historia "Iniciar sesión con Google", la integrante "Carolina" y hoy 18/10/2026

  # Caso NORMAL (CA-24.1, CA-24.3)
  @CA-24.1 @CA-24.3
  Escenario: Registrar horas en la historia
    Cuando "Carolina" registra "3" horas el "2026-10-18" en "Pantalla de login"
    Entonces se registra el esfuerzo
    Y la historia suma "3" horas reales

  # Caso LÍMITE (CA-24.1): de 0 a 24 h por registro
  @CA-24.1
  Esquema del escenario: Horas de un registro
    Cuando "Carolina" registra "<horas>" horas el "2026-10-18" en "Pantalla de login"
    Entonces el registro da como resultado "<resultado>"

    Ejemplos:
      | horas | resultado                                       |
      | 24    | registrado                                      |
      | 0,5   | registrado                                      |
      | 0     | las horas tienen que ser mayores a 0 y hasta 24 |
      | 24,5  | las horas tienen que ser mayores a 0 y hasta 24 |

  # Caso de ERROR (CA-24.1): no más de 24 h en el día
  @CA-24.1
  Esquema del escenario: Total del día
    Dado que "Carolina" ya registró "20" horas el "2026-10-17"
    Cuando "Carolina" registra "<horas>" horas el "2026-10-17" en "Más trabajo"
    Entonces el registro da como resultado "<resultado>"

    Ejemplos:
      | horas | resultado                                     |
      | 4     | registrado                                    |
      | 5     | el integrante no puede superar 24 h en un día |

  # Caso de ERROR (CA-24.2): fechas
  @CA-24.2
  Esquema del escenario: Fecha del registro
    Cuando "Carolina" registra "2" horas el "<fecha>" en "Pantalla de login"
    Entonces el registro da como resultado "<resultado>"

    Ejemplos:
      | fecha      | resultado                                                  |
      | 2026-10-05 | registrado                                                 |
      | 2026-10-19 | la fecha no puede ser futura                               |
      | 2026-10-04 | la fecha tiene que estar dentro de las fechas del proyecto |

  # Caso de ERROR (CA-24.3): el integrante tiene que ser del proyecto
  @CA-24.3
  Escenario: Un integrante dado de baja no registra horas
    Dado que "Carolina" está dada de baja
    Cuando "Carolina" registra "2" horas el "2026-10-18" en "Pantalla de login"
    Entonces el registro da como resultado "el integrante tiene que ser un integrante activo del proyecto"

  # Caso ALTERNATIVO: varias cargas suman en la historia (M5)
  @CA-24.1
  Escenario: Las horas reales suman todos los registros de la historia
    Dado que "Carolina" ya registró "20" horas el "2026-10-17"
    Cuando "Carolina" registra "1,5" horas el "2026-10-18" en "Revisión"
    Entonces la historia suma "21,5" horas reales
