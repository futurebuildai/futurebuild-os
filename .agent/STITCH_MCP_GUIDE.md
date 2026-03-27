# Google Stitch MCP Integration Guide

This guide defines the interface between **Google Stitch** (Generative UI) and the **FutureBuild Agent Kit** via the Model Context Protocol (MCP).

## 1. MCP Connectivity Requirements

To enable Google Stitch to "understand" our design system, the MCP server must provide access to the following project trees:

- **Component Library**: `frontend/src/components/` (To read existing `.ts` and `.css` implementations).
- **Design Tokens**: `frontend/src/styles/theme.css` (To understand glassmorphism and color variables).
- **Asset Directory**: `frontend/public/assets/` (To reference icons and images).

### Connection Config
```json
{
  "mcpServers": {
    "futurebuild-design-hub": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "/home/colton/Desktop/FutureBuild_HQ/dev_ecosystem/futurebuild-repo/frontend/src"]
    }
  }
}
```

## 2. DynamicUIArtifact Schema (Stage 1 Output)

Google Stitch must output UI designs in a standardized JSON format. This allows Antigravity to parse the intent into technical specifications in Stage 2.

### JSON Structure
```json
{
  "artifactType": "DynamicUIArtifact",
  "version": "1.1.0",
  "metadata": {
    "componentName": "ERP_Dashboard_Card",
    "theme": "glass-dark",
    "mobileFirst": true
  },
  "layout": {
    "structure": "flex-col",
    "padding": "var(--spacing-md)",
    "gap": "var(--spacing-sm)"
  },
  "components": [
    {
      "type": "fb-card",
      "props": {
        "variant": "glass",
        "title": "Project Spend",
        "icon": "finance-chart"
      },
      "children": [
        {
          "type": "fb-data-grid",
          "props": {
            "source": "/api/v1/financials/summary",
            "columns": ["Date", "Vendor", "Amount"]
          }
        }
      ]
    }
  ],
  "styles": {
    "customCSS": ".erp-card { backdrop-filter: blur(12px); border: 1px solid var(--glass-border); }"
  }
}
```

## 3. Propagation Protocol (Stage 1 -> Stage 2)

1. **Output**: Stitch writes the above JSON to `.agent/temp/stitch_output.json`.
2. **Trigger**: Antigravity is notified to "Parse Stitch Output".
3. **Action**: Antigravity maps `fb-` components to `frontend/src/components` imports and generates the Lit/TS execution prompt for Claude Code.

## 4. GableLBM Design Constraints

When generating `DynamicUIArtifact` payloads, Google Stitch MUST adhere to the following GableLBM architectural constraints:

1.  **Color Strictness**: Only use `var(--md-sys-color-*)` variables. Never hardcode hex values.
2.  **Component Archetypes**:
    - **Cards**: Must use the `.glass-card` class for layout depth.
    - **Interactions**: Use the `.hover-lift` class for all clickable surface elements.
    - **Loading States**: Define `skeleton` components for any asynchronous data fetching.
3.  **Layout Systems**:
    - Use `var(--fb-spacing-*)` for consistent gaps.
    - Prefer `flex` and `grid` layouts that respect the `12px` (`var(--md-sys-shape-corner-medium)`) corner radius standards.
4.  **Brand Identity**:
    - Primary actions must utilize `var(--md-sys-color-primary)` (Gable Green).
    - Destructive actions must utilize `var(--md-sys-color-error)` (Safety Red).

