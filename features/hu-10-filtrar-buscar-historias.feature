# language: es
@HU-10
Característica: Filtrar y buscar historias
  Como integrante
  Quiero filtrar y buscar historias
  Para encontrarlas rápido

  Antecedentes:
    Dado el proyecto "Software Metrics" con estas historias para filtrar:
      | título              | prioridad | estado      |
      | Iniciar sesión      | Alta      | Pendiente   |
      | Registrar usuario   | Alta      | En sprint   |
      | Editar perfil       | Media     | En progreso |
      | Cerrar sesión       | Baja      | Hecho       |
      | Reporte de sesiones | Media     | Pendiente   |

  # Caso NORMAL (CA-10.1, RN1): sin filtros se ve el backlog completo, en su orden
  @CA-10.1 @RN1
  Escenario: Sin filtros se ven todas las historias
    Cuando filtro el backlog sin filtros
    Entonces veo estas historias:
      | título              |
      | Iniciar sesión      |
      | Registrar usuario   |
      | Editar perfil       |
      | Reporte de sesiones |
      | Cerrar sesión       |

  # Caso NORMAL (CA-10.1): un filtro por vez
  @CA-10.1
  Esquema del escenario: Filtrar por estado o por prioridad
    Cuando filtro el backlog por <filtro> "<valor>"
    Entonces veo estas historias: "<historias>"

    Ejemplos:
      | filtro    | valor       | historias                          |
      | estado    | Pendiente   | Iniciar sesión, Reporte de sesiones |
      | estado    | En progreso | Editar perfil                      |
      | estado    | Hecho       | Cerrar sesión                      |
      | prioridad | Alta        | Iniciar sesión, Registrar usuario  |
      | prioridad | Baja        | Cerrar sesión                      |

  # Caso NORMAL (CA-10.1, RN2): la búsqueda es por el título, sin distinguir mayúsculas ni tildes
  @CA-10.1 @RN2
  Esquema del escenario: Buscar por texto en el título
    Cuando busco "<texto>" en el backlog
    Entonces veo estas historias: "<historias>"

    Ejemplos:
      | texto     | historias                                          |
      | sesión    | Iniciar sesión, Reporte de sesiones, Cerrar sesión |
      | SESION    | Iniciar sesión, Reporte de sesiones, Cerrar sesión |
      | perfil    | Editar perfil                                      |
      | regístrar | Registrar usuario                                  |

  # Caso ALTERNATIVO (CA-10.1, RN1 y RN3): los filtros se combinan y se respeta el orden del backlog
  @CA-10.1 @RN1 @RN3
  Escenario: Combinar los tres filtros
    Cuando filtro el backlog por estado "Pendiente", prioridad "Media" y texto "sesion"
    Entonces veo estas historias:
      | título              |
      | Reporte de sesiones |
    Y se informa que se ven 1 de 5 historias

  # Caso de ERROR (CA-10.2): ninguna historia cumple los filtros
  @CA-10.2
  Escenario: Ninguna historia coincide
    Cuando filtro el backlog por estado "Hecho", prioridad "Alta" y texto ""
    Entonces no veo ninguna historia
    Y se informa "Ninguna historia coincide con los filtros."

  # Caso LÍMITE (RN4): un valor que no existe se ignora
  @CA-10.1 @RN4
  Escenario: Un estado que no existe se ignora
    Cuando filtro el backlog por estado "Archivada"
    Entonces veo 5 historias

  # Caso ALTERNATIVO (RN5): con filtros no se reordena
  @RN5
  Escenario: Con filtros no se ofrecen Subir ni Bajar
    Cuando filtro el backlog por prioridad "Alta"
    Entonces no se ofrece reordenar
    Y se informa "Para cambiar el orden, quitá los filtros."
