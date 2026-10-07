# language: es
@HU-30
Característica: Story points planificados y completados
  Como Agile Enabler
  Quiero ver los story points planificados y completados
  Para medir cuánto cumplimos de lo que nos comprometimos

  # Caso NORMAL (CA-30.1)
  @CA-30.1
  Escenario: Planificados y completados de un sprint
    Dado un sprint con una historia de 5 story points en estado "Hecho"
    Y una historia de 3 story points en estado "Hecho"
    Y una historia de 8 story points en estado "En progreso"
    Cuando consulto los story points del sprint
    Entonces los story points planificados son 16
    Y los story points completados son 8
    Y no se muestra ninguna advertencia de estimación

  # Caso ALTERNATIVO (CA-30.2)
  @CA-30.2
  Escenario: Las historias sin estimar suman 0 y se avisa
    Dado un sprint con una historia de 5 story points en estado "Hecho"
    Y una historia sin estimar en estado "Hecho"
    Y una historia sin estimar en estado "En sprint"
    Cuando consulto los story points del sprint
    Entonces los story points planificados son 5
    Y los story points completados son 5
    Y se muestra la advertencia "hay 2 historias sin estimar"

  # Caso LÍMITE (CA-30.1)
  @CA-30.1
  Escenario: Sprint sin historias
    Dado un sprint sin historias
    Cuando consulto los story points del sprint
    Entonces los story points planificados son 0
    Y los story points completados son 0
    Y no se muestra ninguna advertencia de estimación

  # Caso ALTERNATIVO (CA-30.3)
  @CA-30.3
  Escenario: El total del proyecto es la suma de sus sprints
    Dado un sprint con una historia de 5 story points en estado "Hecho"
    Y una historia de 8 story points en estado "En progreso"
    Y otro sprint con una historia de 3 story points en estado "Hecho"
    Y una historia sin estimar en estado "En sprint"
    Cuando consulto los story points del proyecto
    Entonces los story points planificados son 16
    Y los story points completados son 8
    Y se muestra la advertencia "hay 1 historia sin estimar"

  # Caso de ERROR
  @CA-30.1
  Escenario: Story points negativos
    Dado un sprint con una historia de -3 story points en estado "Hecho"
    Cuando consulto los story points del sprint
    Entonces el cálculo de story points se rechaza con el error "los story points no pueden ser negativos"
