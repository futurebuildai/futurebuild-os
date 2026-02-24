# audit_wal

> **⚠️ SUPERSEDED:** This stub has been replaced by the canonical spec at
> [`internal/audit/wal.md`](file:///home/colton/Desktop/FutureBuild_HQ/XUI/backend/shadow/internal/audit/wal.md).
> See Sprint 3.2: Interactive Learning.

## Intent
*   **High Level:** Write-Ahead Log for user correction events to AI-extracted artifact fields.
*   **Business Value:** Enables VisionAgent prompt refinement and extraction accuracy improvement.

## Responsibility
*   Superseded — see `internal/audit/wal.md` for the full spec including CorrectionEntry struct, AuditWAL methods, and PostgreSQL DDL.

## Dependencies
*   **Upstream:** `internal/api/handlers/correction_handler` → `POST /api/v1/corrections`
*   **Downstream:** PostgreSQL `correction_log` table
