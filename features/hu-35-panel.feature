# language: es
@HU-35
Característica: Panel de indicadores del proyecto
  Como Agile Enabler
  Quiero un dashboard con el estado del proyecto
  Para decidir rápido

  # Caso LÍMITE (CA-35.2): un proyecto sin datos no da error
  @CA-35.2
  Escenario: Panel de un proyecto recién creado
    Dado un proyecto recién creado
    Cuando abro el panel
    Entonces la tarjeta "Estado" muestra "Planificado"
    Y las tarjetas "Sprint activo", "Velocidad", "% completado", "Desviación de horas" y "Defectos abiertos" muestran "sin datos"

  # Caso NORMAL (CA-35.1): cada tarjeta muestra su métrica
  @CA-35.1
  Escenario: Panel con datos
    Dado un proyecto en curso con el Sprint 1 cerrado con 8 SP completados y el Sprint 2 activo
    Y una historia de 5 SP Hecha y otra de 5 SP pendiente
    Y una historia con 10 h estimadas y 12 h trabajadas
    Y 3 defectos detectados y 1 resuelto
    Cuando abro el panel
    Entonces la tarjeta "Estado" muestra "En curso"
    Y la tarjeta "Sprint activo" muestra "Sprint 2"
    Y la tarjeta "Velocidad" muestra "8 SP por sprint"
    Y la tarjeta "% completado" muestra "50,0 %"
    Y la tarjeta "Desviación de horas" muestra "+2 h"
    Y la tarjeta "Defectos abiertos" muestra "2"

  # Caso ALTERNATIVO (CA-35.1): todos los defectos resueltos no es "sin datos"
  @CA-35.1
  Escenario: Todos los defectos resueltos
    Dado un proyecto en curso con el Sprint 1 cerrado con 8 SP completados y el Sprint 2 activo
    Y 1 defectos detectados y 1 resuelto
    Cuando abro el panel
    Entonces la tarjeta "Defectos abiertos" muestra "0"
