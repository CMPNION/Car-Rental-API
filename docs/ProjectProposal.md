# Project Proposal — Car Rental System

## Overview
This project delivers a web application for short‑term vehicle rentals. The product focuses on reliable booking, transparent pricing, and a smooth lifecycle from discovery to payment and completion. The core stack is Go (backend) with REST endpoints and a Nuxt-based client. The main goal is to build a complete, production‑style workflow: authentication, catalog browsing, booking with overlap prevention, payment simulation, and admin operations.

## a) Project relevance (why this topic)
Urban mobility needs flexible, short-term rentals that are simpler than owning a vehicle. Current services often hide availability rules or make pricing hard to predict. This project focuses on predictable scheduling and a secure booking flow that prevents double‑booking. The system also supports a clear financial layer (balance and payment), which is essential for a realistic rental lifecycle.

## b) Competitor analysis (Kazakhstan)
The Kazakh market includes car‑sharing and rental services from local and international brands (e.g., city‑based car‑sharing apps and rental aggregators). These typically provide booking and basic fleet management. Our competitive advantages are:
- **Clear booking safeguards**: Overlap checks with transactional booking and consistent status updates.
- **Transparent status system**: The UI highlights rental states and car availability in a uniform way.
- **Modular backend**: Architecture supports expansion (corporate accounts, discounts, analytics) without major rework.

## c) Target audience
- **Primary**: Students and young professionals needing short‑term rentals.
- **Secondary**: Small businesses or corporate users requiring flexible transport.
- **Admin users**: Fleet operators and support staff managing vehicles and rentals.

## d) Project features
- **Auth & role‑based UI**: JWT token with admin detection.
- **Vehicle catalog**: filtering, sorting, and detail view.
- **Booking engine**: date selection, client‑side price estimate, and server‑side overlap protection.
- **Financial flow**: balance top‑up and payment flow (pending → active).
- **Lifecycle**: cancel/finish operations update rental and car statuses.
- **Admin console**: CRUD for cars and action controls for rentals.

Word count: ~430
