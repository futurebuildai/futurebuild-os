# FutureBuild Product Vision
## Residential Construction Path Model (CPM-res1.0)

**Version:** 1.0.0  
**Classification:** Business & Vision Specification

---

## 1. System Identity

### 1.1 Name
**FutureBuild**

### 1.4 Purpose Statement
FutureBuild automates residential construction scheduling by combining probabilistic AI (document understanding, context extraction) with deterministic algorithms (critical path calculation, duration estimation) to manage construction projects from permit issuance through completion.

### 1.5 Unique Value Proposition
The system treats the PROJECT as the primary entity, not human users. The project is modeled as a living dependency graph that computes its own state, identifies risks, and coordinates stakeholders through AI agents.

---

## 2. Core Philosophy

### 2.1 Project-Native Intelligence
The project is the first-class citizen. Every residential project runs as an independent process with its own task graph, context variables, and schedule state.

### 2.2 Deterministic Core, Probabilistic Perception
| Layer Type | Purpose | Certainty |
|------------|---------|-----------|
| Probabilistic | Understand documents, photos, messages | Variable (0.0-1.0 confidence) |
| Deterministic | Calculate schedules, durations, costs | Exact (reproducible) |

---

## 3. System Boundaries

### 3.1 What FutureBuild Does (In Scope)
- Post-permit scheduling (WBS 5.2+).
- Task dependency management (~80 tasks).
- Duration calculation (DHSM).
- Weather adjustment (SWIM).
- Stakeholder communication & Procurement coordination.

### 3.2 What FutureBuild Does NOT Do (Out of Scope)
- Pre-permit phases.
- Architectural design.
- Payment processing.

---

## 6. Project Lifecycle

### 6.1 Lifecycle Stages
1. **PROJECT CREATION**: Geocoding, record creation.
2. **DOCUMENT PROCESSING**: Extraction of variables via Gemini.
3. **SCHEDULE GENERATION**: DHSM & SWIM application, CPM solver.
4. **ACTIVE CONSTRUCTION**: Morning briefings, real-time updates.
5. **COMPLETION**: Archive, summary report.

---

## 7. Architecture Overview

### 7.1 Layer Stack
- **Layer 0: Real World Inputs** (Documents, Photos)
- **Layer 1: Context Engine** (Gemini AI Extraction)
- **Layer 2: Data Spine** (PostgreSQL)
- **Layer 3: Physics Engine** (CPM/DHSM Calculations)
- **Layer 4: Action Engine** (Agentic Automations)
- **Layer 5: Learning Layer** (Adaptive Multipliers)

---

## 8. Layer 0: Real World Inputs
Handles Blueprint PDFs, site photos, emails, and SMS via webhooks and OCR processing.

---

## 14. User Model
Defined User Types: Owner, Admin, Project Lead, Client, Subcontractor, and Vendor. Access is managed via a Permission Matrix.

---

## 15. Authentication System
Uses a passwordless **Magic Link** flow resulting in restricted JWTs for internal and portal users.

---

## 16. API Specification
RESTful API base: `/api/v1`. Endpoints for Project Management, Scheduling, Tasks, Procurement, Documents, and AI Chat.

---

## 17. Frontend Architecture
Built with VanillaJS, Custom Web Components, and CSS Custom Properties. Features a responsive dashboard with interactive Gantt charts and AI Chat artifacts.

---

## 18. Communication System
Multichannel coordination (Email, SMS, Portal) for task readiness, inspection results, and daily focus briefings.

---

## 23. System Constraints
- Performance targets (<100ms API, <3s Chat).
- Rate limits and 7-year data retention for compliance.

---

## 24. Integration Points
Integrates with Google Vertex AI, SendGrid, Twilio, and various Geocoding/Weather services.
