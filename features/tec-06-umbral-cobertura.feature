# language: es
@TEC-06
Característica: Umbral de cobertura en el CI
  Como equipo
  Quiero que el CI falle si la cobertura del dominio y de las métricas baja del 80 %
  Para no perder sin darnos cuenta la calidad que ya tenemos

  # Caso NORMAL (CA-06.2): la cobertura está muy por encima del mínimo
  @CA-06.2
  Escenario: La cobertura supera el umbral
    Dado que el perfil de cobertura tiene:
      | paquete                 | sentencias | cubiertas |
      | internal/domain/backlog | 141        | 138       |
      | internal/metrics        | 89         | 89        |
    Cuando verifico el umbral de cobertura de 80 %
    Entonces el umbral se cumple
    Y la cobertura combinada es "98,7 %"

  # Caso LÍMITE (RN2): justo en el mínimo todavía pasa
  @CA-06.2 @RN2
  Escenario: La cobertura está justo en el umbral
    Dado que el perfil de cobertura tiene:
      | paquete          | sentencias | cubiertas |
      | internal/metrics | 100        | 80        |
    Cuando verifico el umbral de cobertura de 80 %
    Entonces el umbral se cumple
    Y la cobertura combinada es "80,0 %"

  # Caso de ERROR (RN2): un punto por debajo del mínimo hace fallar el CI
  @CA-06.2 @RN2
  Escenario: La cobertura queda por debajo del umbral
    Dado que el perfil de cobertura tiene:
      | paquete                 | sentencias | cubiertas |
      | internal/domain/backlog | 100        | 79        |
    Cuando verifico el umbral de cobertura de 80 %
    Entonces el umbral no se cumple
    Y el error dice "la cobertura combinada de internal/domain e internal/metrics es 79,0 %, por debajo del mínimo de 80,0 %"

  # Caso ALTERNATIVO (RN1): la cobertura se combina, no se promedia ni se exige a cada paquete
  @CA-06.2 @RN1
  Escenario: La cobertura se combina entre los paquetes
    Dado que el perfil de cobertura tiene:
      | paquete                 | sentencias | cubiertas |
      | internal/domain/backlog | 90         | 90        |
      | internal/metrics        | 10         | 0         |
    Cuando verifico el umbral de cobertura de 80 %
    Entonces el umbral se cumple
    Y la cobertura combinada es "90,0 %"

  # Caso ALTERNATIVO (RN3): lo que no es dominio ni métricas no cuenta
  @CA-06.2 @RN3
  Escenario: Los demás paquetes no cuentan para el umbral
    Dado que el perfil de cobertura tiene:
      | paquete          | sentencias | cubiertas |
      | internal/domain  | 100        | 100       |
      | cmd/server       | 50         | 0         |
      | internal/server  | 200        | 40        |
    Cuando verifico el umbral de cobertura de 80 %
    Entonces el umbral se cumple
    Y el informe tiene una línea para "internal/domain"
    Y el informe no tiene una línea para "cmd/server"
    Y el informe no tiene una línea para "internal/server"

  # Caso LÍMITE (RN4): un bloque repetido en el perfil cuenta una sola vez
  @CA-06.2 @RN4
  Escenario: Un bloque repetido en el perfil cuenta una sola vez
    Dado que el perfil de cobertura repite un bloque de 10 sentencias de "internal/metrics", una vez cubierto y otra no
    Y el perfil de cobertura tiene además:
      | paquete          | sentencias | cubiertas |
      | internal/domain  | 10         | 10        |
    Cuando verifico el umbral de cobertura de 80 %
    Entonces el umbral se cumple
    Y la cobertura combinada es "100,0 %"

  # Caso de ERROR (RN5): si el perfil no trae los paquetes controlados, no se aprueba en silencio
  @CA-06.2 @RN5
  Escenario: El perfil no tiene sentencias de los paquetes controlados
    Dado que el perfil de cobertura tiene:
      | paquete         | sentencias | cubiertas |
      | internal/server | 100        | 100       |
    Cuando verifico el umbral de cobertura de 80 %
    Entonces el umbral no se cumple
    Y el error dice "el perfil no tiene sentencias de los paquetes controlados (internal/domain, internal/metrics): ¿cambió el nombre de algún paquete?"

  # Caso de ERROR (RN6): un perfil que no se puede leer tampoco pasa
  @CA-06.2 @RN6
  Escenario: El perfil está mal formado
    Dado que el perfil de cobertura es un texto que no es un perfil
    Cuando verifico el umbral de cobertura de 80 %
    Entonces el umbral no se cumple
    Y el error empieza con "perfil de cobertura:"
