# MMI Corporate Portal 

A comprehensive Corporate Business Trip (Perjalanan Dinas) and Cash Advance (Realisasi Bon Sementara) Management System. Built with a high-performance Golang backend and a modern React frontend.

## Table of Contents
- [About the Project](#-about-the-project)
- [Core Features](#-core-features)
- [Tech Stack](#-tech-stack)
- [Project Structure](#-project-structure)
- [Getting Started](#-getting-started)
  - [Prerequisites](#prerequisites)
  - [Backend Setup](#backend-setup)
  - [Frontend Setup](#frontend-setup)
  - [Docker Setup](#docker-setup)
- [Usage Workflow](#-usage-workflow)

---

## About the Project
The MMI Corporate Portal digitizes and streamlines internal financial and HR workflows. It provides an end-to-end platform for employees to request business trips (PPD), apply for cash advances, and submit expense realizations. The system enforces a strict multi-tier approval matrix and automatically generates standardized PDF documents for auditing and compliance.

## Core Features
- **Role-Based Access Control (RBAC):** Secure authentication and authorization using JWT, with distinct roles (Pegawai, Atasan, HRGA, Finance, Direktur).
- **Dynamic Approval Workflow:** Automated routing of requests based on the corporate hierarchy (e.g., Employee ➔ Manager ➔ HRGA ➔ Finance ➔ Director).
- **Document Generation:** Automated PDF generation for Business Trip Letters (Surat Perjalanan Dinas) and Cash Advance Receipts using integrated templates.
- **Expense Grouping & Tracking:** Structured grouping of expenses (Accommodation, Transportation, Miscellaneous) with real-time discrepancy calculation (Selisih Bon).
- **Digital Signatures:** Support for capturing and attaching digital signatures (e-Sign) to official documents.
- **Real-Time Notifications:** In-app alerts for pending approvals and document status updates.

## Tech Stack

**Backend**
- **Language:** Go (Golang) 1.21+
- **Framework:** Echo v4
- **ORM:** GORM
- **Database:** PostgreSQL
- **Document Rendering:** GoFPDF / HTML-to-PDF templates

**Frontend**
- **Library:** React 18
- **Language:** TypeScript
- **Build Tool:** Vite
- **Styling:** Tailwind CSS
- **Routing:** React Router DOM
- **HTTP Client:** Axios

**Infrastructure**
- Docker & Docker Compose
- Nginx (Frontend reverse proxy)

---

## Usage Workflow
1. **Authentication & Session Management:**
   - Log in via the `/login` page using your corporate credentials. The system decodes the JWT claims to dynamically grant access permissions based on your specific role (Pegawai, HRGA, Finance, or Direktur).

2. **Business Trip Request (PPD):**
   - Employees initiate a new request by filling out the target destination, purpose, departure/return dates, and structured expense estimates (Accommodation, Transportation, and Miscellaneous).
   - The frontend strictly formats all datetimes into the ISO-8601/RFC3339 standard (`YYYY-MM-DDT00:00:00Z`) before transmission to ensure smooth backend binding.
   - Upon submission, the initial status automatically matches the conditional approval matrix defined for the requester's rank.

3. **Dynamic Approval Matrix (Approval Workflow):**
   - Authorized stakeholders (HRGA, Finance, Direktur) can monitor pending verification requests directly from their dedicated operational dashboards.
   - Approving a request (`Setujui`) records an electronic signature (`e-Sign`) along with a timestamp (`waktu_disetujui`), pushing the document status seamlessly to the next evaluation tier.
   - Declining a request (`Tolak`) requires a mandatory explanation string in the comment field (`catatan`), changing the state directly to `Ditolak`.

4. **Cash Advance Realization (Realisasi Bon Sementara - RBS):**
   - Upon completing the business trip, employees submit actual operational costs backed by digital copies of their receipts.
   - For regular employees (**Pegawai**), the system is optimized to **bypass the immediate Manager (Atasan) and jump straight to HRGA verification** for faster corporate turnaround times.
   - The core engine aggregates actual spending against the allocated cash advance to output a net balance sheet discrepancy (`Selisih Bon`).

5. **Document Finalization & Auditing:**
   - Once the Finance department provides final clearance, the backend triggers automated HTML/GoFPDF templates to map all transactional history, approvals, signatures, and timestamps into an immutable PDF ready for accounting compliance audits.

---

## Project Structure

```text
mmi-project/
├── backend/
│   ├── internal/
│   │   ├── config/        # Database, Auth, and System configuration
│   │   ├── constant/      # Application-wide constants & enums
│   │   ├── dto/           # Data Transfer Objects (Request/Response structures)
│   │   ├── handler/       # HTTP controllers / Route handlers
│   │   ├── middleware/    # JWT Auth, Logger, CORS middlewares
│   │   ├── model/         # GORM Database entities
│   │   ├── repository/    # Database interaction layer
│   │   ├── service/       # Core business logic layer
│   │   └── utils/         # Helper functions (PDF, Password, Formatting)
│   ├── migrations/        # SQL migration & indexing scripts
│   ├── public/            # Static assets (Signatures, Receipts)
│   ├── main.go            # Backend entry point
│   └── dockerfile         # Backend container configuration
├── frontend/
│   ├── src/
│   │   ├── components/    # Reusable UI components (Buttons, Modals, Forms)
│   │   ├── contexts/      # React Context (Global state, Auth state)
│   │   ├── lib/           # Utility libraries (Axios instance, Formatting)
│   │   ├── pages/         # Page components (Dashboard, Login, PPD, Bon)
│   │   └── types/         # TypeScript interface definitions
│   ├── package.json       # Frontend dependencies
│   ├── vite.config.ts     # Vite bundler configuration
│   └── Dockerfile         # Frontend container configuration
└── README.md

