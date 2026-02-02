# Technical Specification: Notification UI (Step 91)

| Metadata | Details |
| :--- | :--- |
| **Step** | 91 |
| **Feature** | Notification Center |
| **Goal** | Implement a bell icon with a dropdown stream of system alerts. |
| **Related** | Phase 15, PRD Section 4.2 |

---

## 1. Feature Description

A centralized location for user notifications (`<fb-notification-bell>`). This resides in the App Shell (Top Bar). When clicked, it opens a drawer or popover containing recent alerts (`<fb-notification-list>`).

### 1.1 Architecture
- **Service**: `NotificationService` (Mocked for now, ready for API).
- **Store**: `NotificationStore` (Signals).
- **UI**: Bell Icon + Popover.

---

## 2. Components

### 2.1 `fb-notification-bell`
**Path**: `frontend/src/components/notifications/fb-notification-bell.ts`

- **Visual**: Bell Icon (Lucide).
- **Badge**: Red circle with number if `unread > 0`.
- **Interaction**: Click toggles the Popover.

### 2.2 `fb-notification-list`
**Path**: `frontend/src/components/notifications/fb-notification-list.ts`

- **Visual**: Vertical list of items.
- **Item Content**:
    - **Icon**: Info (Blue), Alert (Red), Success (Green).
    - **Title**: Bold text.
    - **Time**: Relative time (e.g., "5m ago").
    - **Action**: Click navigates to link.

### 2.3 `NotificationService`
**Path**: `frontend/src/services/notification-service.ts`

- **Method**: `getNotifications()` - returns `Observable<Notification[]>`.
- **Method**: `markAsRead(id)` - updates state locally.
- **Mock Data**:
    ```typescript
    const MOCK_NOTIFICATIONS = [
      { id: '1', type: 'alert', title: 'Schedule Slip', message: 'Project Alpha is 3 days behind.' },
      { id: '2', type: 'success', title: 'Build Complete', message: 'Deployment #123 successful.' }
    ];
    ```

---

## 3. Implementation Steps

### Step 3.1: Create Service
- Implement `NotificationService` with mock data.
- Expose a `notifications$` signal or observable.

### Step 3.2: Create `fb-notification-bell`
- Standard Lit component.
- Subscribe to `notifications$`.
- Render badge conditionally.

### Step 3.3: Create `fb-notification-list`
- Render the list logic.
- Style with CSS Grid/Flexbox.

### Step 3.4: Integrate into App Shell
- Add `<fb-notification-bell>` to `fb-app.ts` header section.

---

## 4. Verification Plan

### 4.1 Automated Browser Testing (Claude in Chrome)

**CRITICAL INSTRUCTION**: You must use the `/chome` extension (or equivalent Browser Tool) to verify this feature.

**Workflow**:
1. **Launch Browser**: Open `http://localhost:8080`.
2. **Locate Element**: Find the Bell Icon in the top right header.
3. **Verify Badge**:
    - Ensure the red badge is visible (given mock data has 2 unread items).
    - Read the badge text (should be "2").
4. **Interaction**:
    - Click the Bell Icon.
    - **Verify**: A dropdown/popover appears.
    - **Verify Content**: Ensure "Schedule Slip" text is visible in the list.
5. **Dismissal (Optional)**:
    - Click a "Mark Read" button if implemented.
    - Verify badge count decreases.

**Auto-Accept**:
- If using `/chome`, assume **Auto-Accept** permissions for localhost testing.
