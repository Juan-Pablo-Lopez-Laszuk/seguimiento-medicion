# language: es
#
# PLANTILLA BDD — Software Metrics & Estimation
# ------------------------------------------------------------
# Cómo usarla:
#   1. Copiá este archivo con el nombre de tu historia: features/hu-XX-nombre.feature
#   2. Reemplazá lo que está entre [corchetes] y borrá estos comentarios.
#      (Los <nombres> entre < > sí se dejan: son las columnas de la tabla Ejemplos.)
#   3. Escribí al menos un escenario por cada tipo de caso que pide la consigna:
#      normal, alternativo, límite y error.
#   4. Cada escenario lleva la etiqueta de su criterio de aceptación (@CA-XX.n):
#      así se sigue el hilo Historia → Spec → Criterio → Escenario → Test → Código.
#
# Palabras clave en español (las entiende godog gracias a "# language: es"):
#   Característica  = la historia de usuario
#   Antecedentes    = pasos que se repiten al inicio de todos los escenarios
#   Escenario       = un ejemplo concreto de comportamiento
#   Esquema del escenario + Ejemplos = el mismo escenario con varias filas de datos
#   Dado / Cuando / Entonces / Y / Pero = contexto / acción / resultado esperado
# ------------------------------------------------------------

@HU-XX
Característica: [Título de la historia]
  Como [rol]
  Quiero [objetivo]
  Para [beneficio]

  Antecedentes:
    Dado un proyecto "[nombre del proyecto]"

  # Caso NORMAL: el camino feliz
  @CA-XX.1
  Escenario: [el comportamiento esperado en la situación habitual]
    Dado [el contexto inicial]
    Cuando [la acción del usuario]
    Entonces [el resultado esperado]

  # Caso ALTERNATIVO: otra situación válida
  @CA-XX.2
  Escenario: [una variante válida]
    Dado [otro contexto]
    Cuando [la acción del usuario]
    Entonces [el resultado esperado]

  # Caso LÍMITE: varios datos de prueba en una tabla
  @CA-XX.3
  Esquema del escenario: [el comportamiento en los bordes]
    Dado [un contexto] con el valor "<dato>"
    Cuando [la acción del usuario]
    Entonces el resultado es "<esperado>"

    Ejemplos:
      | dato | esperado |
      | 0    | 0        |
      | 1    | 1        |

  # Caso de ERROR: datos inválidos
  @CA-XX.4
  Escenario: [el sistema rechaza un dato inválido]
    Dado [un contexto]
    Cuando [la acción con un dato inválido]
    Entonces se informa el error "[mensaje de error]"
