# go-record-signer
A message-driven microservice for signing records in concurrent batches using Ed25519, built with Go, NATS JetStream, and PostgreSQL.

## Quickstart

The project uses a Makefile to simplify running different components:

```bash
# Initialize the database with records and keys
make init

# Optional: Run dispatcher only to prepare batches in NATS queue
make dispatch

# Run both dispatcher and two worker instances in parallel
make sign

# Check the current database status (signed vs unsigned records)
make check

# To restart the process with clean data
make init
```

## Project Overview

This project implements a record signing service using a message-driven microservice architecture. It handles concurrent signing of records using a pool of private keys, ensuring that each record is signed exactly once and no key is used simultaneously by multiple processes.

### Components

#### initdb

The initialization component that:
- Sets up database schema
- Generates the specified number of Ed25519 key pairs (default: 100)
- Encrypts private keys with AES-GCM before storing them (for simplicity, private keys are stored in the database encrypted)
- Creates unsigned records with random data (default: 100,000)
- Stores everything in PostgreSQL database

#### dispatcher

The dispatcher service that:
- Retrieves batches of unsigned records from the database
- Creates message batches and publishes them to NATS JetStream
- Updates record status from PENDING to QUEUED
- Continues until all records are queued or an error occurs

#### worker

The worker service that:
- Subscribes to the record batches queue in NATS
- Acquires a signing key using the least-recently-used (LRU) strategy
- Signs all records in a batch with the same key
- Updates the database with signatures and record status
- Ensures no key is used concurrently by multiple workers

## Implementation Notes

The project focuses on simplicity while meeting the core requirements. Some areas that could be improved in a production environment:

- **Logging**: Currently using basic log package; could be enhanced with structured logging
- **Error handling**: Basic error handling is implemented without sophisticated retry mechanisms
- **Configuration**: Uses simple environment variables instead of a more robust configuration system
- **Metrics and monitoring**: No metrics collection or health endpoints
- **Testing**: Has unit tests for crypto functions, but could benefit from integration tests
- **Graceful shutdown**: Basic cleanup implemented, but lacks comprehensive graceful shutdown
- **Key management**: For simplicity, private keys are stored encrypted in the database; a more secure approach would use an HSM, vault service, or key management system in production
- **Worker implementation**: For simplicity, concurrency is achieved by running multiple worker instances. An alternative approach could use Go's concurrency features (goroutines) within a single worker process

## Original Task

**Goal:**  
Implement a **record signing service** using a **message-driven/microservice architecture**.

**Scenario:**  
Given:
- A database with **100,000 records**
- A collection of **100 private keys**

Create a process to **concurrently sign batches of records**, storing the resulting **signatures** in the database until all records have been signed.

**Rules:**
1. **No double signing:**  
   Each record must be signed only once.
2. **No concurrent usage of keys:**  
   A private key must not be used simultaneously in multiple processes.
3. **Batching:**  
   - All records in a batch should be signed with the **same key**.  
   - Select keys based on **least recently used** (LRU) strategy.  
   - The **batch size** should be **configurable** at the start (not changeable at runtime).
