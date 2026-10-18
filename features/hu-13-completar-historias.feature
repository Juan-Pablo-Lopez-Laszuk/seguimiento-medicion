# language: es
@HU-13
Característica: Registrar historias completadas
  Como Product Builder
  Quiero marcar una historia del sprint como completada
  Para registrar el avance

  Antecedentes:
    Dado el proyecto "Software Metrics" con el Sprint 2 activo del "2026-10-12" al "2026-10-18"
    Y que el Sprint 3 empieza el "2026-10-19"
    Y el Sprint 2 tiene estas historias para completar:
      | historia      | sp | criterios cumplidos | estado      |
      | HU-1 Login    | 5  | 2 de 2              | En progreso |
      | HU-2 Registro | 3  | 1 de 2              | En progreso |
      | HU-3 Perfil   | 5  | 1 de 1              | En sprint   |

  # Caso NORMAL: se completa y se guarda la fecha (CA-13.1 y CA-13.2)
  @CA-13.1 @CA-13.2
  Escenario: Completar una historia del sprint activo
    Dado que hoy es el "2026-10-15" para completar
    Cuando completo "HU-1"
    Entonces "HU-1" queda Hecha con fecha de completado "2026-10-15"
    Y el Sprint 2 lleva 5 de 13 SP completados

  # Caso de ERROR: faltan criterios por cumplir (CA-13.3)
  @CA-13.3
  Escenario: Faltan criterios
    Dado que hoy es el "2026-10-15" para completar
    Cuando completo "HU-2"
    Entonces completar da el error "HU-2 no se puede completar: falta cumplir 1 de 2 criterios"
    Y "HU-2" sigue "En progreso"

  # Caso de ERROR: hay que pasar por En progreso (CA-08.1, decisión 6)
  @CA-13.3
  Escenario: Hay que empezarla antes
    Dado que hoy es el "2026-10-15" para completar
    Cuando completo "HU-3"
    Entonces completar da el error "HU-3 todavía no empezó: primero hay que pasarla a En progreso"
    Y "HU-3" sigue "En sprint"

  # Caso de ERROR: solo el sprint activo (CA-13.1)
  @CA-13.1
  Escenario: No se completa fuera del sprint activo
    Dado que hoy es el "2026-10-15" para completar
    Y que "HU-4 Ayuda" está En progreso en el Sprint 3 con sus criterios cumplidos
    Cuando completo "HU-4"
    Entonces completar da el error "solo se completan historias del sprint activo"

  # Caso LÍMITE: la fecha es la de Argentina aunque en UTC ya sea el día siguiente (CA-13.2)
  @CA-13.2
  Escenario: La fecha es la de Argentina
    Dado que ahora son las "22:00" del "2026-10-18" en Argentina
    Cuando completo "HU-1"
    Entonces "HU-1" queda Hecha con fecha de completado "2026-10-18"

  # Caso LÍMITE: un sprint vencido sin cerrar sigue activo (HU-04 RN2)
  @CA-13.1
  Escenario: Sprint vencido sin cerrar
    Dado que hoy es el "2026-10-20" para completar
    Y que el Sprint 2 todavía no se cerró
    Cuando completo "HU-1"
    Entonces "HU-1" queda Hecha con fecha de completado "2026-10-20"
