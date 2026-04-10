# AP2 Assignment 1

## Clean Architecture Microservices (Order & Payment)

### Overview

This project implements a small microservice platform written in Go.

* **Order Service** – manages customer orders
* **Payment Service** – processes payments

The services communicate using **REST over HTTP** and each service follows **Clean Architecture principles**.

Each service has its own database and business logic.

---

# Architecture

Each service is structured using Clean Architecture layers:

* **Domain** – entities and business rules
* **Use Case** – application logic
* **Repository** – database access
* **Transport (HTTP)** – REST handlers
* **App / main.go** – dependency injection and service initialization

Handlers are thin and only process HTTP requests and responses.
All business logic is implemented inside the use case layer.

---

# Bounded Contexts

The system is divided into two independent bounded contexts.

### Order Service

Responsible for:

* creating orders
* storing order information
* updating order status

The order lifecycle can be:

Pending → Paid / Failed / Cancelled

The Order Service calls the Payment Service to validate payments.

### Payment Service

Responsible for:

* authorizing payments
* generating transaction IDs
* enforcing payment limits

The Payment Service stores all payment transactions in its own database.

Each service has its **own database and models**.
There is **no shared code or shared database** between services.

---

# Failure Handling

The Order Service calls the Payment Service using an HTTP client with a **2 second timeout**.

If the Payment Service is unavailable:

1. The request times out.
2. The order status is updated to **Failed**.
3. The API returns **503 Service Unavailable**.

This prevents the system from waiting indefinitely and ensures the order state is consistent.

---

# Business Rules

* Order amount must be **greater than 0**
* Money values are stored as **int64 (cents)** to avoid floating point errors
* **Paid orders cannot be cancelled**
* If payment amount **> 100000 cents**, the payment is **Declined**

---

# API Examples

### Create Order

POST /orders

Example request:

curl -X POST http://localhost:8080/orders
-H "Content-Type: application/json"
-d '{"customer_id":"cust1","item_name":"Laptop","amount":50000}'

---

### Get Order

GET /orders/{id}

curl http://localhost:8080/orders/{id}

---

### Cancel Order

PATCH /orders/{id}/cancel

curl -X PATCH http://localhost:8080/orders/{id}/cancel

---

### Create Payment

POST /payments

curl -X POST http://localhost:8081/payments
-H "Content-Type: application/json"
-d '{"order_id":"order1","amount":50000}'

---

### Get Payment

GET /payments/{order_id}

curl http://localhost:8081/payments/{order_id}

---

# Architecture Decisions

### Why Clean Architecture

Clean Architecture separates business logic from infrastructure code.
This makes the system easier to test and maintain.

### Why int64 for money

Floating point numbers can cause rounding errors.
Using int64 cents ensures financial accuracy.

### Why separate databases

Each service owns its data to maintain clear service boundaries and avoid tight coupling.

---

# Bonus

The system implements **Idempotency-Key** support for creating orders.
Sending the same request with the same key will return the existing order instead of creating a duplicate.
