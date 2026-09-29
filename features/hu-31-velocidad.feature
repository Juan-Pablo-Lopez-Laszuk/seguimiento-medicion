# language: es
@HU-31
Característica: Velocidad del equipo
  Como Agile Enabler
  Quiero conocer la velocidad del equipo
  Para decidir cuántos story points comprometer en el próximo sprint

  # Caso NORMAL y ALTERNATIVO (CA-31.1 y CA-31.2)
  @CA-31.1 @CA-31.2
  Esquema del escenario: Promedio de los últimos sprints cerrados
    Dado un proyecto con una ventana de velocidad de <ventana> sprints
    Y sprints cerrados que completaron "<completados>" story points
    Cuando consulto la velocidad del equipo
    Entonces la velocidad informada es <velocidad>

    Ejemplos:
      | ventana | completados | velocidad |
      | 3       | 10,20,30,40 | 30        |
      | 3       | 10,20       | 15        |
      | 1       | 12,8        | 8         |

  # Caso LÍMITE (CA-31.3)
  @CA-31.3
  Escenario: Proyecto sin sprints cerrados
    Dado un proyecto con una ventana de velocidad de 3 sprints
    Y que el proyecto no tiene sprints cerrados
    Cuando consulto la velocidad del equipo
    Entonces la velocidad informada es 0
    Y se muestra el mensaje "Aún no hay sprints cerrados"

  # Caso ALTERNATIVO (CA-31.4)
  @CA-31.4
  Escenario: Los sprints activos o planificados no cuentan
    Dado un proyecto con una ventana de velocidad de 3 sprints
    Y sprints cerrados que completaron "10,20" story points
    Y un sprint activo que lleva completados 50 story points
    Cuando consulto la velocidad del equipo
    Entonces la velocidad informada es 15

  # Caso de ERROR (CA-31.5)
  @CA-31.5
  Esquema del escenario: Ventana de cálculo inválida
    Dado un proyecto con una ventana de velocidad de <ventana> sprints
    Cuando consulto la velocidad del equipo
    Entonces se informa el error "la ventana de sprints debe ser mayor a cero"

    Ejemplos:
      | ventana |
      | 0       |
      | -1      |
