# language: es
@TEC-04
Característica: Chequeo de salud de la app publicada
  Como equipo
  Quiero que /health diga si la base de datos responde
  Para saber si el deploy en Vercel quedó bien configurado

  # Caso NORMAL (CA-04.3): la app guarda en la base y la base responde
  @CA-04.3
  Escenario: La base de datos responde
    Dado que la app tiene una base de datos que responde
    Cuando consulto el estado de la app
    Entonces la respuesta de salud tiene el código 200
    Y el estado general es "ok"
    Y el estado de la base es "ok"

  # Caso ALTERNATIVO (RN1): sin DATABASE_URL la app guarda en memoria y lo avisa
  @CA-04.3
  Escenario: La app corre sin base de datos
    Dado que la app corre sin base de datos
    Cuando consulto el estado de la app
    Entonces la respuesta de salud tiene el código 200
    Y el estado general es "ok"
    Y el estado de la base es "memoria"

  # Caso de ERROR (RN2 y RN4): la base no responde y el error no se muestra
  @CA-04.3
  Escenario: La base de datos no responde
    Dado que la app tiene una base de datos que no responde
    Cuando consulto el estado de la app
    Entonces la respuesta de salud tiene el código 503
    Y el estado general es "error"
    Y el estado de la base es "sin conexión"

  # Caso LÍMITE (RN3): se espera a la base como máximo 3 segundos
  @CA-04.3
  Esquema del escenario: La base de datos tarda en responder
    Dado que la app tiene una base de datos que tarda <segundos> segundos en responder
    Cuando consulto el estado de la app
    Entonces la respuesta de salud tiene el código <codigo>
    Y el estado de la base es "<base>"

    Ejemplos:
      | segundos | codigo | base         |
      | 2        | 200    | ok           |
      | 4        | 503    | sin conexión |
