# E-commerce Backend (Go + Gin + GORM + PostgreSQL)

## Description
This is a **multi-vendor e-commerce backend** built using **Golang (Gin + GORM)** with **PostgreSQL** as the database.  
It supports role-based access with **SuperAdmin, Admin, Vendor, and User** roles, product management, vendor approval workflows, and more.

## Features
-  Authentication & Role-based Authorization
-  Vendor Approval Flow (Pending → Approved/Rejected/Suspended)
-  Product Management with Status (`draft`, `archive`, `private`, `published`)
-  Product Filtering based on Role:
  - Public users → only see `published`
  - Vendors → manage their own products
  - Admin/SuperAdmin → manage all products
-  Category Management
-  Audit Logs for important actions

## Tech Stack
- **Backend:** Go (Gin, GORM)
- **Database:** PostgreSQL
- **Frontend (planned):** React + TypeScript
- **Deployment (planned):** Docker, Netlify (for frontend)

## Installation
1. Clone the repository:
   ```bash
   git clone https://github.com/abdullahalsazib/e-com-backend.git
2. Navigate to the project folder:
    ```bash
      cd e-com-backend
3. Set up your .env file with database and JWT configuration.
4. Run the server:
     ```bash
     go run main.go
