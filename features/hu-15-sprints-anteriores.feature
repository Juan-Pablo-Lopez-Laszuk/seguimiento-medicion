# language: es
@HU-15
Característica: Consultar sprints anteriores
  Como integrante
  Quiero consultar los sprints anteriores
  Para analizar cómo trabajamos

  Antecedentes:
    Dado el proyecto "Software Metrics" con estos sprints para consultar:
      | sprint | goal                 | inicio     | fin        |
      | 1      | Gestionar el backlog | 2026-10-12 | 2026-10-18 |
      | 2      | Planificar sprints   | 2026-10-19 | 2026-10-25 |
      | 3      | Métricas             | 2026-10-26 | 2026-11-01 |
    Y que el Sprint 1 se cerró con estas historias:
      | historia      | sp | completada |
      | HU-1 Login    | 5  | sí         |
      | HU-2 Registro | 3  | no         |
      | HU-3 Perfil   | 5  | no         |
    Y que el Sprint 2 se cerró con estas historias:
      | historia      | sp | completada |
      | HU-4 Tablero  | 3  | sí         |
      | HU-5 Reportes | 3  | sí         |
      | HU-6 Exportar | 2  | no         |

  # Caso NORMAL: la lista con el resultado de cada sprint cerrado, ordenada por número (CA-15.1)
  @CA-15.1
  Escenario: Ver los sprints anteriores con su resultado
    Cuando consulto los sprints del proyecto
    Entonces la lista de sprints muestra, en este orden:
      | sprint   | goal                 | inicio     | fin        | sp completados | cumplimiento |
      | Sprint 1 | Gestionar el backlog | 12/10/2026 | 18/10/2026 | 5 de 13        | 38,5 %       |
      | Sprint 2 | Planificar sprints   | 19/10/2026 | 25/10/2026 | 6 de 8         | 75,0 %       |
      | Sprint 3 | Métricas             | 26/10/2026 | 01/11/2026 | —              | —            |

  # Caso NORMAL: el detalle muestra la foto del cierre (CA-15.2)
  @CA-15.2
  Escenario: Ver el detalle de un sprint cerrado
    Cuando consulto el detalle del Sprint 1
    Entonces el detalle dice "5 de 13 SP completados · 38,5 % de cumplimiento · 1 de 3 historias"
    Y el detalle muestra estas historias al cerrar:
      | historia | título   | sp | resultado     |
      | HU-1     | Login    | 5  | Completada    |
      | HU-2     | Registro | 3  | No completada |
      | HU-3     | Perfil   | 5  | No completada |

  # Caso ALTERNATIVO: el detalle no cambia con lo que pasa después del cierre (RN6)
  @CA-15.2
  Escenario: El detalle conserva la foto aunque la historia cambie después
    Dado que "HU-2" se re-estimó en 8 story points y entró al Sprint 3
    Cuando consulto el detalle del Sprint 1
    Entonces el detalle muestra "HU-2" con 3 SP y "No completada"
    Y el detalle dice "5 de 13 SP completados · 38,5 % de cumplimiento · 1 de 3 historias"

  # Caso LÍMITE: un sprint cerrado sin historias no divide por cero (RN3)
  @CA-15.1
  Escenario: Sprint cerrado sin historias
    Dado que el Sprint 3 se cerró sin historias
    Cuando consulto los sprints del proyecto
    Entonces la lista muestra el Sprint 3 con "0 de 0" y cumplimiento "—"
    Y el detalle del Sprint 3 dice "El sprint se cerró sin historias"

  # Caso LÍMITE: todo completado (RN2)
  @CA-15.1
  Escenario: Sprint cerrado con todas sus historias completadas
    Dado que el Sprint 3 se cerró con estas historias:
      | historia     | sp | completada |
      | HU-7 Alertas | 5  | sí         |
    Cuando consulto los sprints del proyecto
    Entonces la lista muestra el Sprint 3 con "5 de 5" y cumplimiento "100,0 %"

  # Caso de ERROR: el sprint no existe (CA-15.2)
  @CA-15.2
  Escenario: Consultar un sprint que no existe
    Cuando consulto el detalle del Sprint 9
    Entonces la consulta del sprint responde "No encontramos ese sprint."
