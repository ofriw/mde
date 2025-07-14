# 2025-07-14 - [FEATURE] Development Approach
**Priority**: High

Overall development strategy for MDE project.

## Architecture Decisions
- **Internal plugins only** for MVP - compiled into binary for simplicity
- **Elm Architecture** via Bubble Tea for predictable state management  
- **Single flat Model** approach as recommended by Bubble Tea best practices
- **goldmark** for markdown parsing - extensible and CommonMark compliant

## Development Principles
- **Incremental delivery** - working editor at each milestone
- **Test-driven** - 80% coverage target with testify/gomock
- **Performance first** - profile early and often
- **LLM-friendly** - clear interfaces and minimal code

## Tech Stack Rationale
- **Go**: Performance, single binary distribution
- **Bubble Tea**: Modern TUI framework with good ecosystem
- **goldmark**: Most popular Go markdown parser, extensible
- **Cobra/Viper**: Standard Go CLI tooling
- **Lip Gloss**: Pairs well with Bubble Tea for styling

## Risk Mitigations
- Start with terminal compatibility testing early
- Keep plugin interfaces simple for v1
- Focus on core editing experience before advanced features
- Use VHS for visual regression testing