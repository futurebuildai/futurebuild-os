# Technical Specification: Mobile Navigation (Step 90)

| Metadata | Details |
| :--- | :--- |
| **Step** | 90 |
| **Feature** | Mobile Navigation Bar |
| **Goal** | Implement a bottom tab bar for screens < 768px to improve mobile usability. |
| **Related** | Phase 15, PRD Section 4.1 |

---

## 1. Feature Description

The current sidebar navigation is not suitable for mobile devices. We will implement a "Bottom Tab Bar" (`<fb-mobile-nav>`) that serves as the primary navigation on small screens, hiding the sidebar.

### 1.1 Visual Spec
- **Height**: 64px
- **Position**: Fixed Bottom (`bottom: 0`, `left: 0`, `right: 0`)
- **Background**: `var(--surface-color)` (Glassmorphism if possible)
- **Border**: Top border `1px solid var(--border-color)`
- **Items**: 4 Tabs (Dashboard, Projects, Chat, Menu)
- **Active State**: Icon changes color to Primary, Label bold/visible.

---

## 2. Architecture & Components

### 2.1 New Component: `fb-mobile-nav`
**Path**: `frontend/src/components/layout/fb-mobile-nav.ts`

```typescript
@customElement('fb-mobile-nav')
export class MobileNav extends LitElement {
  // Simple router-link based navigation
  // State: Active tab derived from Router.location
}
```

### 2.2 App Shell Updates
**Path**: `frontend/src/fb-app.ts`

We need to conditionally render `fb-mobile-nav` and toggle the visibility of `fb-sidebar` based on media queries.

**CSS Strategy**:
```css
/* In fb-app.css or inline styles */
@media (max-width: 767px) {
  fb-sidebar { display: none; }
  fb-mobile-nav { display: flex; }
  main { padding-bottom: 80px; } /* Prevent content being covered */
}
@media (min-width: 768px) {
  fb-mobile-nav { display: none; }
}
```

---

## 3. Implementation Steps

### Step 3.1: Create `fb-mobile-nav.ts`
- Import standard icons from `lucide-lit` or similar (Home, Folder, MessageSquare, Menu).
- Implement standard Lit boilerplates.
- Use `<a>` tags with `class="nav-item"` for routing to `/`, `/projects`, `/chat`.

### Step 3.2: Update `fb-app.ts`
- Import `fb-mobile-nav`.
- Add `<fb-mobile-nav></fb-mobile-nav>` to the DOM, adjacent to `<fb-sidebar>`.
- Add the media query CSS to handle display toggling.

### Step 3.3: "Menu" Drawer (Optional for MVP)
- The 4th tab "Menu" should simply dispatch an event `toggle-mobile-menu` or route to `/settings` for this iteration. Let's route to `/settings` for simplicity, labeling it "Settings".

---

## 4. Verification Plan

### 4.1 Automated Browser Testing (Claude in Chrome)

**CRITICAL INSTRUCTION**: You must use the `/chome` extension (or equivalent Browser Tool) to verify this feature visually.

**Workflow**:
1. **Launch Browser**: Open the dev server URL (e.g., `http://localhost:8080`).
2. **Set Viewport**: Resize the browser window to **375px wide** (Mobile/iPhone size).
3. **Verify Constraints**:
    - Confirm `<fb-sidebar>` is **NOT** visible.
    - Confirm `<fb-mobile-nav>` **IS** visible at the bottom.
4. **Interactive Test**:
    - Click "Projects" tab.
    - Verify URL changes to `/projects`.
    - Verify the "Projects" icons highlights as active.
5. **Desktop Regression**:
    - Resize window back to **1024px**.
    - Confirm `<fb-mobile-nav>` disappears and `<fb-sidebar>` reappears.

**Auto-Accept**:
- If using `/chome` or similar tool that prompts for permission, assume **Auto-Accept** permissions for localhost testing.

### 4.2 Manual Verification
- Open on actual mobile device if available (via local network IP).
- Check for "Safe Area" interference on iOS (bottom bar overlap).

---

## 5. Guardrails
- **No JavaScript Media Queries**: Use CSS media queries for visibility to avoid FOUC (Flash of Unstyled Content).
- **Z-Index**: Ensure nav bar is `z-index: 1000` to sit above content but below modals.
- **Touch Targets**: Tap targets must be at least 44x44px.
