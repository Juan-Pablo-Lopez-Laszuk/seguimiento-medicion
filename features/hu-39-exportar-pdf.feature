# language: es
@HU-39
Característica: Exportar reporte a PDF
  Como Agile Enabler
  Quiero exportar el reporte a PDF
  Para compartirlo fuera del sistema

  Antecedentes:
    Dado el proyecto "Software Metrics" con estos sprints para el PDF:
      | sprint | goal                 | inicio     | fin        |
      | 1      | Gestionar el backlog | 2026-10-12 | 2026-10-18 |
      | 2      | Planificar sprints   | 2026-10-19 | 2026-10-25 |
      | 3      | Ver métricas         | 2026-10-26 | 2026-11-01 |
    Y estas historias del proyecto para el PDF:
      | historia      | sp | horas estimadas | horas reales |
      | HU-1 Login    | 5  | 6               | 8            |
      | HU-2 Registro | 3  | 4               | 3            |
      | HU-3 Perfil   | 5  | 2               | 0            |
    Y que para el PDF el Sprint 1 se cerró con "HU-1, HU-2, HU-3" y completó "HU-1"
    Y que para el PDF el Sprint 2 se cerró con "HU-2, HU-3" y completó "HU-2"
    Y que para el PDF "HU-3" está en el Sprint 3
    Y que el PDF se genera el 29/10/2026

  # Caso NORMAL: el PDF de un sprint se descarga con el nombre del proyecto, el sprint y la fecha (CA-39.1)
  @CA-39.1
  Escenario: Descargar el PDF del reporte de un sprint
    Cuando descargo el PDF del reporte del Sprint 2
    Entonces se descarga el archivo "software-metrics-sprint-2-2026-10-29.pdf"

  # Caso NORMAL: el PDF tiene lo mismo que el reporte en pantalla (CA-39.2, RN1)
  @CA-39.2
  Escenario: El PDF de un sprint coincide con su reporte
    Cuando descargo el PDF del reporte del Sprint 2
    Entonces el PDF muestra lo mismo que el reporte del Sprint 2 en pantalla

  # Caso ALTERNATIVO: un sprint abierto lleva el aviso de reporte parcial (RN2)
  @CA-39.2
  Escenario: El PDF de un sprint abierto avisa que es parcial
    Cuando descargo el PDF del reporte del Sprint 3
    Entonces el PDF dice "El sprint no está cerrado: el reporte muestra cómo va hasta hoy."

  # Caso NORMAL: el PDF del proyecto, con un bloque por sprint igual a su reporte (CA-39.1, CA-39.2)
  @CA-39.1 @CA-39.2
  Escenario: Descargar el PDF del reporte del proyecto
    Cuando descargo el PDF del reporte del proyecto
    Entonces se descarga el archivo "software-metrics-2026-10-29.pdf"
    Y el PDF muestra lo mismo que el reporte del proyecto en pantalla
    Y cada sprint del PDF empieza en una hoja nueva

  # Caso LÍMITE: un proyecto grande se genera dentro del límite de Vercel (CA-39.3)
  @CA-39.3
  Escenario: El PDF de un proyecto grande se genera a tiempo
    Dado el proyecto "Proyecto grande" con 12 sprints cerrados de 30 historias cada uno
    Cuando descargo el PDF del reporte del proyecto "Proyecto grande"
    Entonces el PDF se genera en menos de 1 segundo

  # Caso LÍMITE: caracteres que las fuentes del PDF no tienen (RN4)
  @CA-39.2
  Escenario: Un título con caracteres que el PDF no puede mostrar
    Dado que "HU-3" se llama "Perfil 🔐 → OAuth"
    Cuando descargo el PDF del reporte del Sprint 3
    Entonces el PDF muestra la historia "Perfil ? ? OAuth"

  # Caso de ERROR: el sprint no existe
  @CA-39.1
  Escenario: Pedir el PDF de un sprint que no existe
    Cuando descargo el PDF del reporte del Sprint 9
    Entonces la descarga responde 404 con "No encontramos ese sprint."
