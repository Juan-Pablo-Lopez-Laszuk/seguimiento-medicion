# language: es
@HU-12
Característica: Asignar historias a un sprint
  Como equipo
  Quiero asignar historias al sprint
  Para comprometer el trabajo de la iteración

  Antecedentes:
    Dado el proyecto "Software Metrics" para planificar sprints
    Y que el Sprint 1 está cerrado con 16 SP completados
    Y que el Sprint 2 es el sprint activo
    Y que el Sprint 3 todavía no empezó
    Y el backlog para asignar tiene estas historias:
      | historia       | sp          | criterios | estado    |
      | HU-1 Login     | 5           | 2         | Pendiente |
      | HU-2 Registro  | sin estimar | 1         | Pendiente |
      | HU-3 Perfil    | 3           | 0         | Pendiente |
      | HU-4 Reportes  | 0           | 1         | Pendiente |
      | HU-5 Ayuda     | 2           | 1         | Hecho     |

  # Caso NORMAL (CA-12.1)
  @CA-12.1
  Escenario: Asignar una historia lista
    Cuando asigno "HU-1" al Sprint 2
    Entonces "HU-1" queda en el Sprint 2 en estado "En sprint"
    Y el Sprint 2 suma 5 SP comprometidos

  # Caso LÍMITE: 0 story points es una estimación (CA-12.1)
  @CA-12.1
  Escenario: Una historia de 0 SP se puede asignar
    Cuando asigno "HU-4" al Sprint 2
    Entonces "HU-4" queda en el Sprint 2 en estado "En sprint"

  # Casos de ERROR: historias que no están listas (CA-12.1)
  @CA-12.1
  Esquema del escenario: Historias que no se pueden asignar
    Cuando asigno "<historia>" al Sprint 2
    Entonces la asignación da el error "<error>"
    Y "<historia>" sigue sin sprint

    Ejemplos:
      | historia | error                                                       |
      | HU-2     | HU-2 no se puede asignar: no está estimada                  |
      | HU-3     | HU-3 no se puede asignar: no tiene criterios de aceptación  |
      | HU-5     | HU-5 no se puede asignar: ya está Hecha                     |

  # Caso de ERROR: sprint cerrado (CA-12.2)
  @CA-12.2
  Escenario: No se asigna a un sprint cerrado
    Cuando asigno "HU-1" al Sprint 1
    Entonces la asignación da el error "el Sprint 1 está cerrado: no se le pueden asignar historias"

  # Caso de ERROR: una historia en un solo sprint abierto (CA-12.2)
  @CA-12.2
  Escenario: La historia ya está en otro sprint
    Dado que asigné "HU-1" al Sprint 2
    Cuando asigno "HU-1" al Sprint 3
    Entonces la asignación da el error "HU-1 ya está en el Sprint 2"

  # Caso ALTERNATIVO: planificar el sprint siguiente antes de que empiece (RN2)
  @CA-12.2
  Escenario: Planificar el sprint siguiente
    Cuando asigno "HU-1" al Sprint 3
    Entonces "HU-1" queda en el Sprint 3 en estado "En sprint"

  # Caso ALTERNATIVO: quitar una historia que todavía no empezó (RN4)
  @CA-12.2
  Escenario: Quitar una historia del sprint
    Dado que asigné "HU-1" al Sprint 2
    Cuando quito "HU-1" del sprint
    Entonces "HU-1" sigue sin sprint
    Y el Sprint 2 suma 0 SP comprometidos

  # Caso NORMAL y LÍMITE: aviso por velocidad, sin bloquear (CA-12.3)
  @CA-12.3
  Escenario: Aviso por velocidad
    Dado que el Sprint 2 ya tiene 13 SP comprometidos
    Cuando asigno "HU-1" al Sprint 2
    Entonces "HU-1" queda en el Sprint 2 en estado "En sprint"
    Y se avisa "Ojo: el Sprint 2 suma 18 SP y la velocidad del equipo es 16,0 SP"

  # Caso LÍMITE: sin sprints cerrados no hay velocidad y no se avisa (CA-12.3, decisión 8)
  @CA-12.3
  Escenario: Sin sprints cerrados no hay aviso
    Dado que el proyecto no tiene sprints cerrados para comparar
    Y que el Sprint 2 ya tiene 40 SP comprometidos
    Cuando asigno "HU-1" al Sprint 2
    Entonces no hay aviso de velocidad
