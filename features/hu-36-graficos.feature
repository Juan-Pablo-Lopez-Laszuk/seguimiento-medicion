# language: es
@HU-36
Característica: Gráficos de métricas
  Como equipo
  Quiero ver gráficos de las métricas
  Para ver tendencias

  Antecedentes:
    Dado el proyecto "Tienda" con estos sprints para los gráficos:
      | sprint | objetivo | inicio     | fin        |
      | 1      | Login    | 2026-10-05 | 2026-10-11 |
      | 2      | Carrito  | 2026-10-12 | 2026-10-18 |
    Y para los gráficos estas historias:
      | historia     | sprint | SP | completada el | horas estimadas | horas reales |
      | HU-1 Login   | 1      | 5  | 2026-10-08    | 6               | 7            |
      | HU-2 Menú    | 1      | 3  | 2026-10-10    | 4               | 5            |
      | HU-3 Perfil  | 1      | 5  |               | 0               | 0            |
      | HU-4 Carrito | 2      | 8  | 2026-10-13    | 8               | 5            |
      | HU-5 Pago    | 2      | 2  | 2026-10-15    | 0               | 0            |
      | HU-6 Envío   | 2      | 3  |               | 0               | 0            |
    Y que el Sprint 1 se cerró para los gráficos
    Y para los gráficos estos defectos:
      | defecto         | severidad | detectado en | resuelto en |
      | Login lento     | Alta      | 1            | 1           |
      | Botón roto      | Crítica   | 2            |             |
      | Texto cortado   | Baja      | 2            |             |
      | Error de cálculo | Alta     | 2            |             |
    Y hoy es 2026-10-15 para los gráficos

  # Caso NORMAL (CA-36.1): velocidad y story points por sprint, de la foto del cierre y de hoy
  @CA-36.1
  Escenario: Velocidad y story points por sprint
    Cuando pido los datos de los gráficos
    Entonces el gráfico de story points tiene:
      | sprint   | planificados | completados |
      | Sprint 1 | 13           | 8           |
      | Sprint 2 | 13           | 10          |
    Y el gráfico de velocidad tiene:
      | sprint   | completados | velocidad |
      | Sprint 1 | 8           | 8         |

  # Caso NORMAL (CA-36.1): horas y defectos
  @CA-36.1
  Escenario: Horas por sprint y defectos por severidad
    Cuando pido los datos de los gráficos
    Entonces el gráfico de horas tiene:
      | sprint   | estimadas | reales |
      | Sprint 1 | 10        | 12     |
      | Sprint 2 | 8         | 5      |
    Y el gráfico de defectos tiene:
      | severidad | detectados | abiertos |
      | Crítica   | 1          | 1        |
      | Alta      | 2          | 1        |
      | Media     | 0          | 0        |
      | Baja      | 1          | 1        |

  # Caso NORMAL (CA-36.1): burndown del sprint activo; los días que no pasaron no tienen dato
  @CA-36.1
  Escenario: Burndown del sprint activo
    Cuando pido los datos de los gráficos
    Entonces el burndown es del "Sprint 2" con:
      | día   | ideal | pendientes |
      | 12/10 | 13    | 13         |
      | 13/10 | 10,8  | 5          |
      | 14/10 | 8,7   | 5          |
      | 15/10 | 6,5   | 3          |
      | 16/10 | 4,3   | —          |
      | 17/10 | 2,2   | —          |
      | 18/10 | 0     | —          |

  # Caso ALTERNATIVO (CA-36.1): el burndown de un sprint cerrado, con la foto del cierre
  @CA-36.1
  Escenario: Burndown de un sprint cerrado
    Cuando pido los datos de los gráficos del Sprint 1
    Entonces el burndown es del "Sprint 1" con:
      | día   | ideal | pendientes |
      | 05/10 | 13    | 13         |
      | 06/10 | 10,8  | 13         |
      | 07/10 | 8,7   | 13         |
      | 08/10 | 6,5   | 8          |
      | 09/10 | 4,3   | 8          |
      | 10/10 | 2,2   | 5          |
      | 11/10 | 0     | 5          |

  # Caso NORMAL (CA-36.2): los datos salen de un endpoint JSON
  @CA-36.2
  Escenario: Los datos salen en JSON
    Cuando pido el JSON de los gráficos
    Entonces el JSON responde 200 con "velocidad", "storyPoints", "horas", "defectos" y "burndown"
    Y el JSON dice que el burndown es del "Sprint 2" y su último día no tiene SP pendientes

  # Caso LÍMITE: un proyecto sin sprints no tiene burndown
  @CA-36.1 @CA-36.2
  Escenario: Proyecto sin sprints
    Dado el proyecto "Vacío" sin sprints para los gráficos
    Cuando pido el JSON de los gráficos
    Entonces el JSON responde 200 con "velocidad", "storyPoints", "horas", "defectos" y "burndown"
    Y el JSON no tiene burndown

  # Caso de ERROR (CA-36.2): un proyecto que no existe
  @CA-36.2
  Escenario: Proyecto inexistente
    Cuando pido el JSON de los gráficos del proyecto 99
    Entonces el JSON responde 404 con el error "No encontramos ese proyecto."
