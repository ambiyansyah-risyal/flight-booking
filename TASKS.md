# Flight Booking - Task Breakdown

This document lists all the tasks for the Flight Booking & Management System based on the case study.

## Task Priority Matrix

| Tasks                        | Description                                                                                                                                                                      |
| ---------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| **Project Setup**            | Set up project structure (CLI/GUI). Make Dockerfile and docker-compose.yml. Define database tables (airport, airports, routes, bookings, passengers)            |
| **Airplane Management**      | Add airplane (ID and seat number). Basic add, read, update, delete for airplane                                                                                                |
| **Airport Management**       | Add airports (city/code). Basic add, read, update, delete for airports                                                                                                                  |
| **Route & Schedule Management**   | Make routes between airports. Assign airport to routes. Schedule flights for certain days                                                                       |
| **Booking (Direct)**         | Search flights between airports. Auto-assign seat and confirm booking. Make booking confirmation with unique ID                                                      |
| **Booking (Transit)**        | Find connecting routes (start → middle → end). Check seats available for all parts. Reserve seats for all legs                                               |
| **Seat Management**          | Update seat numbers after booking or cancelling. Release seats correctly. Stop overbooking                                                                          |
| **Cancellation Management**       | Cancel direct booking. Cancel connecting booking and release seats                                                                                                           |
| **Flight Execution**         | Move days forward (calendar). Stop booking 1 day before flight. Update flight status (scheduled → left → arrived). Handle connecting flights |
| **Passenger Tracking**       | Track passenger status (booked, on board, arrived). Handle special cases (missed connection, full flight)                                                                      |
| **User Interface**           | CLI commands (Admin: manage system; Passenger: book/cancel/view). GUI (optional)                                                                                        |
| **Documentation & Delivery** | Write README and design notes. Instructions for `docker-compose up`. Push to Git (lunch and end of day)                                                                        |

---

**Note**: This task list follows the original case study and focuses on making a strong system with good testing and docs.
