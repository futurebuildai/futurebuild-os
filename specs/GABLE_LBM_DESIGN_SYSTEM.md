# GableLBM Design System Specification

This document defines the mapping and integration of the **GableLBM** design language into the FutureBuild ERP frontend architecture.

## 1. Core Color Palette (HSL & Hex)

The following tokens must be updated in `frontend/src/styles/variables.css` to match the GableLBM "Industrial Dark" aesthetic.

| Token | GableLBM Value (HSL) | Hex Equivalent | Description |
| :--- | :--- | :--- | :--- |
| `--md-sys-color-background` | `230 20% 5%` | `#0A0B10` | Deep Space (Base surface) |
| `--md-sys-color-primary` | `158 100% 50%` | `#00FFA3` | Gable Green (Accent) |
| `--md-sys-color-secondary` | `217 33% 17%` | `#161821` | Slate Steel (Surface 1) |
| `--md-sys-color-tertiary` | `199 95% 60%` | `#38BDF8` | Blueprint Blue |
| `--md-sys-color-error` | `346 87% 60%` | `#F43F5E` | Safety Red |
| `--md-sys-color-surface-container` | `217 33% 17%` | `#161821` | Main Card Background |
| `--md-sys-color-surface-container-high`| `229 19% 11%`| `#161821` | Alternative Elevated Surface |

## 2. Utility Class Overrides (`FBElement.ts`)

The base utilities in `frontend/src/components/base/FBElement.ts` must be updated to match GableLBM's precise glassmorphism and animation settings.

### .glass-card
- **Background**: `rgba(22, 24, 33, 0.6)` (Slate Steel at 60% opacity)
- **Blur**: `24px` (backdrop-filter)
- **Border**: `1px solid rgba(255, 255, 255, 0.05)` (White at 5% opacity)
- **Shadow**: Use `--md-sys-elevation-1`

### .hover-lift
- **Transition**: `transform 0.3s ease-out, box-shadow 0.3s ease-out`
- **Hover Transform**: `translateY(-4px)` (Equivalent to Tailwind `-translate-y-1`)
- **Hover Shadow**: Use `--md-sys-elevation-2`

### .skeleton
- **Animation**: `shimmer 2s linear infinite`
- **Gradient**: 
  ```css
  linear-gradient(
    90deg,
    var(--md-sys-color-surface-container-high) 25%,
    var(--md-sys-color-surface-container) 50%,
    var(--md-sys-color-surface-container-high) 75%
  )
  ```

## 3. Typography Scale

GableLBM utilizes **Outfit** for sans-serif and **JetBrains Mono** for data.

- **Standard Radius**: `0.75rem` (12px) for cards and modals.
- **Large Radius**: `1.5rem` (24px) for specific containers.

## 4. Implementation Instructions for Claude Code

1.  **Overwrite** `frontend/src/styles/variables.css` with the mapped GableLBM hex values.
2.  **Update** `FBElement.ts` static `styles` property with the new utility specifications defined above.
3.  **Run** `npm --prefix frontend run lint` to ensure no syntax errors.
