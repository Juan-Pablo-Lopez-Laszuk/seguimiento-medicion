## Qué resuelve

Closes #<!-- número del issue -->

<!-- Resumen en 1 o 2 oraciones de lo que cambia. -->

## Trazabilidad

| Artefacto | Enlace |
|---|---|
| Historia | HU-XX |
| Especificación SDD | `specs/HU-XX-....md` |
| Escenarios BDD | `features/hu-xx-....feature` |
| Tests | `internal/.../xxx_test.go` |

## Checklist (Definition of Done)

- [ ] La spec SDD está completa y actualizada.
- [ ] Los escenarios BDD cubren caso normal, alternativo, límite y error.
- [ ] Los tests se escribieron primero (hay commits `test(...): RED` antes de `feat(...): GREEN`).
- [ ] `go test ./...` pasa localmente.
- [ ] `go vet ./...` y el linter no reportan problemas.
- [ ] Probé la funcionalidad en el preview de Vercel y cumple los criterios de aceptación.
- [ ] Si usé IA, lo registré en `docs/ia/registro.md`.

## Cómo probarlo

<!-- Pasos para que el revisor verifique el cambio. -->
