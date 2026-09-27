# Definition of Ready y Definition of Done

## Definition of Ready — una historia puede entrar al sprint si:

- [ ] Está escrita en formato **Como / Quiero / Para**.
- [ ] Tiene **criterios de aceptación** acordados con el cliente.
- [ ] Está **estimada** en story points (Planning Poker).
- [ ] Sus **dependencias** están identificadas.
- [ ] Tiene **responsable** asignado.

## Definition of Done — una historia está terminada si:

- [ ] La **especificación SDD** está completa en `specs/` y fue revisada.
- [ ] Los **escenarios BDD** (normal, alternativo, límite y error) están en `features/` y pasan.
- [ ] Los **tests unitarios** se escribieron con TDD: el historial muestra commits RED → GREEN → REFACTOR.
- [ ] La **cobertura** de los paquetes de dominio y métricas es de al menos 80 %.
- [ ] `go vet` y el linter no reportan problemas.
- [ ] El **Pull Request** fue revisado y aprobado por otro integrante.
- [ ] Está **desplegada en Vercel** y se verificaron los criterios de aceptación en la aplicación.
- [ ] Si se usó IA, quedó registrado en `docs/ia/registro.md`.
