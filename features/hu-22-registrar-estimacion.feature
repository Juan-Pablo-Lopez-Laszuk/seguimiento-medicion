# language: es
@HU-22
Característica: Registrar la estimación acordada
  Como Agile Enabler
  Quiero registrar el valor acordado
  Para que quede como estimación de la historia

  Antecedentes:
    Dado el proyecto "Software Metrics" con estos integrantes para Planning Poker:
      | nombre   | rol            | activo |
      | Juan     | AgileEnabler   | sí     |
      | Carolina | ProductBuilder | sí     |
      | Mariano  | ProductBuilder | sí     |
    Y estas historias para estimar:
      | título   | estado    |
      | Login    | Pendiente |
      | Registro | En sprint |
    Y que "Juan" abrió una sesión de Planning Poker para "Login"

  # Caso NORMAL (CA-22.2, CA-22.3): con la ronda revelada, el Agile Enabler registra el valor y la sesión se cierra
  @CA-22.2 @CA-22.3
  Escenario: El Agile Enabler registra la estimación acordada
    Dado que se reveló la ronda actual de "Login"
    Cuando "Juan" registra 5 como estimación acordada de "Login"
    Entonces la historia "Login" tiene 5 story points
    Y la historia "Login" no tiene una sesión de Planning Poker abierta
    Y la historia "Login" tiene estas sesiones cerradas:
      | sesión   | rondas | valor acordado |
      | Sesión 1 | 1      | 5              |

  # Caso ALTERNATIVO (CA-22.2, RN2): re-estimar con Planning Poker anota el cambio en el historial de HU-16b
  @CA-22.2 @RN2
  Escenario: Re-estimar con Planning Poker queda en el historial
    Dado que la historia "Registro" está estimada en 3 story points
    Y que "Juan" abrió una sesión de Planning Poker para "Registro"
    Y que se reveló la ronda actual de "Registro"
    Cuando "Juan" registra 8 como estimación acordada de "Registro"
    Entonces la historia "Registro" tiene 8 story points
    Y el historial de estimaciones de "Registro" tiene el cambio de 3 a 8

  # Caso LÍMITE (CA-22.2, RN2): el mismo valor cierra la sesión pero no anota nada
  @CA-22.2 @RN2
  Escenario: Acordar el mismo valor cierra la sesión sin anotar nada
    Dado que la historia "Registro" está estimada en 3 story points
    Y que "Juan" abrió una sesión de Planning Poker para "Registro"
    Y que se reveló la ronda actual de "Registro"
    Cuando "Juan" registra 3 como estimación acordada de "Registro"
    Entonces la historia "Registro" tiene 3 story points
    Y el historial de estimaciones de "Registro" está vacío
    Y la historia "Registro" no tiene una sesión de Planning Poker abierta

  # Caso LÍMITE (CA-22.1): 0 es un valor de la escala
  @CA-22.1
  Escenario: Cero es una estimación
    Dado que se reveló la ronda actual de "Login"
    Cuando "Juan" registra 0 como estimación acordada de "Login"
    Entonces la historia "Login" tiene 0 story points

  # Caso de ERROR (CA-22.1): un valor fuera de la escala no se registra y la sesión sigue abierta
  @CA-22.1
  Escenario: Un valor fuera de la escala no se registra
    Dado que se reveló la ronda actual de "Login"
    Cuando "Juan" registra 4 como estimación acordada de "Login"
    Entonces registrar la estimación da el error "el valor tiene que ser uno de la escala"
    Y la historia "Login" no tiene story points
    Y la historia "Login" tiene una sesión de Planning Poker abierta en la ronda 1

  # Caso de ERROR (RN4): sin revelar la ronda no se acuerda
  @RN4
  Escenario: No se registra con la ronda sin revelar
    Cuando "Juan" registra 5 como estimación acordada de "Login"
    Entonces registrar la estimación da el error "para registrar la estimación, primero revelá la ronda"
    Y la historia "Login" no tiene story points

  # Caso de ERROR (RN3): solo el Agile Enabler registra
  @RN3
  Escenario: Un Product Builder no registra la estimación
    Dado que se reveló la ronda actual de "Login"
    Cuando "Carolina" registra 5 como estimación acordada de "Login"
    Entonces registrar la estimación da el error "solo el Agile Enabler del proyecto registra la estimación acordada"
    Y la historia "Login" tiene una sesión de Planning Poker abierta en la ronda 1

  # Caso de ERROR (CA-22.3, RN5): una sesión cerrada no se vuelve a cerrar
  @CA-22.3 @RN5
  Escenario: No se registra dos veces
    Dado que se reveló la ronda actual de "Login"
    Y que "Juan" registró 5 como estimación acordada de "Login"
    Cuando "Juan" registra 8 como estimación acordada de "Login"
    Entonces registrar la estimación da el error "esta historia no tiene una sesión de Planning Poker abierta"
    Y la historia "Login" tiene 5 story points

  # Caso LÍMITE (CA-22.3, RN5): las sesiones cerradas quedan para consultar y se puede abrir otra
  @CA-22.3 @RN5
  Escenario: Las sesiones cerradas se consultan desde la historia
    Dado que se reveló la ronda actual de "Login"
    Y que "Juan" registró 5 como estimación acordada de "Login"
    Y que "Juan" abrió una sesión de Planning Poker para "Login"
    Y que se reveló la ronda actual de "Login"
    Cuando "Juan" registra 8 como estimación acordada de "Login"
    Entonces la historia "Login" tiene estas sesiones cerradas:
      | sesión   | rondas | valor acordado |
      | Sesión 2 | 1      | 8              |
      | Sesión 1 | 1      | 5              |

  # Caso de ERROR (RN7): una historia Hecha no se estima, aunque tenga la sesión abierta
  @RN7
  Escenario: Una historia que pasó a Hecha no se estima
    Dado que la historia "Registro" está estimada en 3 story points
    Y que "Juan" abrió una sesión de Planning Poker para "Registro"
    Y que se reveló la ronda actual de "Registro"
    Y que la historia "Registro" pasó a Hecha
    Cuando "Juan" registra 5 como estimación acordada de "Registro"
    Entonces registrar la estimación da el error "la historia está Hecha: no se puede cambiar su estimación"
    Y la historia "Registro" tiene 3 story points
    Y la historia "Registro" tiene una sesión de Planning Poker abierta en la ronda 1
