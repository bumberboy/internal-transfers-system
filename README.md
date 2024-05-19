# Internal Transfers System

This is a simple backend service to showcase financial transactions between accounts.

### Requirements

- **Go**: Ensure you have Go v1.22 installed.
- **Docker**: Docker is required to run PostgreSQL in a container.
- **Make**: Make is required to use the provided Makefile commands. You can also just copy and paste the commands from the Makefile.
- **Postman** or **cURL**: This is for manually testing the endpoints.

### How to run
0. Ensure you have the tools listed in Requirements installed.
1. **Run a plain PostgreSQL Database**:
    ```shell
    make run-db
    ```

2. **Apply Database Migrations**:
    ```shell
   # this might fail if the db is not ready. wait a few seconds and try again.
    make migrate-db
    ```

3. **Run the Application**:
    ```shell
    make run
    ```

### How to manually test endpoints:

1. Ensure app is running

2. Open the `openapi.yaml` file in postman and set your baseUrl variable to localhost:8080. Or just run the following commands below:

3. Create account A:
    ```shell
    curl -i -X POST http://localhost:8080/accounts \
      -H "Content-Type: application/json" \
      -d '{
            "account_id": 8,
            "initial_balance": "1000"
          }' \
      -w "\nHTTP Status: %{http_code}\n"
    ```
4. Create account B:
    ```shell
    curl -i -X POST http://localhost:8080/accounts \
      -H "Content-Type: application/json" \
      -d '{
            "account_id": 9,
            "initial_balance": "0"
          }' \
      -w "\nHTTP Status: %{http_code}\n"
    ```
5. Create a new transfer from account A to B for `500.999999999`:
    ```shell
    curl -i -X POST http://localhost:8080/transactions \
      -H "Content-Type: application/json" \
      -d '{
            "source_account_id": 8,
            "destination_account_id": 9,
            "amount": "500.999999999"
          }' \
      -w "\nHTTP Status: %{http_code}\n"
    ```
6. Verify that account A's balance is now `499.000000001`:
    ```shell
    curl -i -X GET http://localhost:8080/accounts/8 \
      -w "\nHTTP Status: %{http_code}\n"
    ```
### How to run tests
```sh
# if db is already running and already has data in there:
make remove-db # this will remove the existing db along with its data

# start db and init schema
make run-db
make migrate-db # can fail if db is not ready. Wait a few seconds and try again.

make test
```
- This will run both unit and integration tests and generate a coverage report.


## Design

### Fiber as web framework
Fiber was chosen as the web framework for this project due to its high performance and ease of use, inspired by Expressjs. However, this is not compatible with the standard `net/http` package in Go and this means we lose support for libraries in the `net/http` ecosystem. I thought this was a fair tradeoff for a take home assignment.

### Layered architecture

This project follows a simple layered architecture to promote separation of concerns, improved maintainability, code reusability and facilitate testing.  The 2 primary layers are:
1. Presentation layer: handles routing, request validation, response formatting.
2. Service layer: contains business logic, handles data access.

As the project grows, the data access logic can be separated from the service layer into its own layer.

Such an approach makes it easy for application functions in the service layer to be reused and called from other sources such as a CLI command, from an AWS Lambda, or a message queue process: the application logic is independent of how the request/response is processed.

### GORM as ORM
GORM is chosen for its ease of use and extensive feature set. GORM simplifies database operations by providing an intuitive API for common tasks such as CRUD operations, transactions, and migrations, and it allows you to quickly break out into writing raw SQLs. I know devs can get quite opinionated around ORMs. Having used a few ORMs, this is understandable. Not all ORMs are designed properly.

### Decimal
Floating-point arithmetic can lead to precision errors, which are unacceptable in financial applications. The `github.com/shopspring/decimal` library was chosen to handle monetary values accurately as golang does not have first class support for decimals.

The `NUMERIC` postgresql data type is used in the db, with 78 significant digits and 18 decimal places (`NUMERIC(78,18)`). This is enough to store `uint256`, which ETH, denominated in wei, is usually stored in.

Computations in such precision is usually much slower and should be reserved for when it is absolutely necessary like in a withdrawal, deposit, or transfer. For display purposes, operations can be done in float64.

### Passing Context
Context is used to manage request-scoped values, deadlines, cancellation signals, and other request-related data. By passing context to database operations and other long-running tasks, the application can handle timeouts and cancellations effectively. This approach improves the robustness and responsiveness of the system, especially under high load or when interacting with external services.

When collecting traces, context is also used to carry baggage or contextual information that can be useful for debugging errors.

### Test Design
I've written both unit and integration tests for this project.

- **Unit Tests**: Focus on individual components in isolation, such as functions and methods, to verify their behavior under various conditions. See `validator/validators_test.go`. 

- **Integration Tests**: There's 3 integration test suites:
  - `test/integration_test.go`: simple endpoint tests for the 3 endpoints
  - `test/transfer_test.go`: covers some edge cases for the "create transaction" endpoint
  - `test/concurrent_transfer_test.go`: test concurrent transfers for typical concurrency issues. Refer below for how deadlocks/lock contention is mitigated.

You can run the tests with `make test`. The integration tests will require a live postgresql db to run successfully.

### Lock Contention
Lock contention occurs when multiple transactions attempt to acquire locks on the same resources simultaneously, leading to delays and potential deadlocks.
I've employed optimistic concurrency control to reduce lock contention.

In this implementation (in `service/transfer.go: ProcessTransfer`), within a transaction, we first read the rows for the source and destination accounts.
We then validate the transfer amount and shift the amount over.
Before updating the accounts with the new balances, we check that the `updated_at` values are the same as when we first read the records.
If they are different, the transaction terminates.
There's a backoff retry logic that will re-attempt the transaction.

## Improvements
### Potential Problems (and solutions) with existing design

#### Handling Lock Contention
While optimistic concurrency control helps reduce lock contention, there are still potential issues, especially under high transaction loads. The primary challenges include:
- **High Retry Overhead**: Under heavy load, frequent retries can lead to increased latency and higher overall system load.
- **Read-Modify-Write Race Conditions**: Even with optimistic locking, there is a small window between reading and updating records where other transactions might make changes, necessitating retries.
- **Database Load**: High contention scenarios can result in excessive retries, which can increase the load on the database and impact performance.

To further improve our ability to handle these scenarios, we can consider the following strategies:

- **Partitioning**: Split large accounts into multiple smaller accounts. For example, in an incentives center, a marketing account used for distributing reward money can be partitioned into multiple sub-accounts. This reduces contention when many users redeem rewards simultaneously.
- **Batching**: Combine multiple transactions into a single batch operation. This approach reduces the number of locks required, thereby minimizing contention.
- **Application-Level Queuing**: Implement an application-level queuing system to serialize access to the same set of accounts. This ensures that transactions involving the same accounts are processed in sequence, reducing the likelihood of conflicts. We can queue on just the source account or just the destination account or both. 

Each of these suggestions have their pros and cons and trade-offs have to be made based on our specific requirements. Different account usage patterns will likely benefit from different optimisations.

#### Use of uint64 for IDs
The use of `uint64` as IDs for accounts and transfers can be problematic for systems interacting with JavaScript, which has a precision limit of 53 bits for integers. We might need to convert this to string before responding to ensure max compatibility.

### Additional Improvements
#### API Spec: 
Accurate API documentation is crucial for providing a good developer experience (DX). It serves as the foundation of trust and reliability for developers integrating with our API.
There are 3 ways I've seen teams document APIs:
  1. Maintain 2 sources of truth: in code and in an API spec (sometimes in openAPI, sometimes not. There's a good chance they diverge.)
  2. Generate openAPI spec from annotated code. There are some languages/frameworks that can do this very elegantly through some form of reflection/metaprogramming.
  3. API-First Design: Write the openAPI spec first, then generate server stubs and clients. The openAPI spec is the source of truth for what APIs the service offers.

  Due to the time constraint, and the small number of APIs, the openAPI spec for this project was written separately from the code. In an actual project, the API-first design is recommended to ensure consistency and to allow parallel development workflows.

#### protobuf vs OpenAPI:
  - This project's APIs were documented according to openAPI specifications as the assignment requested for an HTTP service. 
  - Nowadays, however, I would recommend the use of `protobuf` for API specification, `buf` for code generation and `connectrpc` as a network protocol in a new project.
  - Protobuf is a much more exact and concise schema definition language than OpenAPI. 
  - You get a more refined set of primitives (`double`, `uint64`, `bytes`, etc.). This reduces effort on conversions and evaluations.
  - Proto is easier to write than yamls (for openapi).
  - Protobuf serialized binary representation is more space efficient than json's ASCII encoding. Request and response payloads are much smaller.
  - Protobuf serialization/deserialization is cheaper than JSON decoding/encoding.
  - `buf` has great documentation and DX. Generation of server stubs and clients in different programming languages (typescript, swift, kotlin, etc.) is easy.
  - `connectrpc` supports the full grpc specification, including client, server or duplex streaming, even across the web.
