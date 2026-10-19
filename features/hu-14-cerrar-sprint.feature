# language: es
@HU-14
Característica: Cerrar sprint
  Como Agile Enabler
  Quiero cerrar el sprint
  Para consolidar sus resultados

  Antecedentes:
    Dado el proyecto "Software Metrics" con el Sprint 1 del "2026-10-12" al "2026-10-18" para cerrar
    Y que el Sprint 2 va del "2026-10-19" al "2026-10-25"
    Y el Sprint 1 tiene estas historias al cerrar:
      | historia      | sp | estado      |
      | HU-1 Login    | 5  | Hecho       |
      | HU-2 Registro | 3  | En progreso |
      | HU-3 Perfil   | 5  | En sprint   |

  # Caso NORMAL: se guarda la foto con los SP planificados y completados (CA-14.3)
  @CA-14.3
  Escenario: Cerrar el sprint activo guarda su resultado
    Dado que hoy es el "2026-10-18" para cerrar
    Cuando cierro el Sprint 1
    Entonces el Sprint 1 queda Cerrado con 13 SP planificados y 5 SP completados
    Y la foto del Sprint 1 tiene estas historias:
      | historia | sp | completada |
      | HU-1     | 5  | sí         |
      | HU-2     | 3  | no         |
      | HU-3     | 5  | no         |

  # Caso NORMAL: las no completadas vuelven al backlog y las Hechas se quedan (CA-14.2)
  @CA-14.2
  Escenario: Las historias no completadas vuelven al backlog
    Dado que hoy es el "2026-10-18" para cerrar
    Cuando cierro el Sprint 1
    Entonces "HU-2" queda "Pendiente" y sin sprint
    Y "HU-3" queda "Pendiente" y sin sprint
    Y "HU-1" sigue "Hecho" en el Sprint 1

  # Caso ALTERNATIVO: la foto no cambia aunque después se re-estime una historia (CA-14.3)
  @CA-14.3
  Escenario: La foto no cambia después del cierre
    Dado que hoy es el "2026-10-18" para cerrar
    Y que cerré el Sprint 1
    Cuando re-estimo "HU-2" en 8 story points
    Entonces la foto del Sprint 1 sigue con 3 SP para "HU-2"
    Y el Sprint 1 sigue con 13 SP planificados

  # Caso ALTERNATIVO: la velocidad sale del sprint cerrado (HU-31)
  @CA-14.3
  Escenario: La velocidad usa el sprint cerrado
    Dado que hoy es el "2026-10-18" para cerrar
    Cuando cierro el Sprint 1
    Entonces la velocidad del equipo es de 5 SP

  # Caso LÍMITE: un sprint vencido sin cerrar sigue activo y se puede cerrar tarde (CA-14.1, HU-04 RN2)
  @CA-14.1
  Escenario: Cerrar un sprint vencido
    Dado que hoy es el "2026-10-20" para cerrar
    Cuando cierro el Sprint 1
    Entonces el Sprint 1 queda Cerrado con 13 SP planificados y 5 SP completados

  # Caso LÍMITE: cerrar antes de la fecha de fin vale, con aviso (RN2)
  @CA-14.1
  Escenario: Cerrar antes de la fecha de fin
    Dado que hoy es el "2026-10-16" para cerrar
    Cuando pido cerrar el Sprint 1
    Entonces la confirmación avisa "El sprint termina el 18/10/2026: lo estás cerrando 2 días antes"
    Y la confirmación dice que vuelven al backlog "HU-2" y "HU-3"

  # Caso LÍMITE: un sprint sin historias se cierra con 0 SP (RN7)
  @CA-14.3
  Escenario: Sprint sin historias
    Dado que hoy es el "2026-10-18" para cerrar
    Y que el Sprint 1 no tiene historias
    Cuando cierro el Sprint 1
    Entonces el Sprint 1 queda Cerrado con 0 SP planificados y 0 SP completados

  # Caso de ERROR: un sprint futuro no se cierra (CA-14.1)
  @CA-14.1
  Escenario: No se cierra un sprint que no es el activo
    Dado que hoy es el "2026-10-18" para cerrar
    Cuando cierro el Sprint 2
    Entonces cerrar da el error "solo se cierra el sprint activo (el Sprint 1)"
    Y el Sprint 2 sigue sin cerrar

  # Caso de ERROR: un sprint cerrado no se cierra de nuevo (CA-14.4)
  @CA-14.4
  Escenario: Un sprint cerrado no se cierra otra vez
    Dado que hoy es el "2026-10-18" para cerrar
    Y que cerré el Sprint 1
    Cuando cierro el Sprint 1
    Entonces cerrar da el error "el Sprint 1 ya está cerrado"

  # Caso de ERROR: un sprint cerrado no se modifica (CA-14.4)
  @CA-14.4
  Escenario: A un sprint cerrado no se le asignan historias
    Dado que hoy es el "2026-10-18" para cerrar
    Y que cerré el Sprint 1
    Cuando asigno "HU-2" al Sprint 1 cerrado
    Entonces asignar da el error "el Sprint 1 está cerrado: no se le pueden asignar historias"

  # Caso de ERROR: sin sprint activo no se cierra nada (CA-14.1)
  @CA-14.1
  Escenario: Hoy no hay sprint activo
    Dado que hoy es el "2026-10-10" para cerrar
    Cuando cierro el Sprint 1
    Entonces cerrar da el error "solo se cierra el sprint activo: hoy no hay ninguno"
