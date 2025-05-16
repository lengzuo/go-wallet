# Wallets

For this project, i will use docker to containerize the application include database, redis, and application. This project is intentional design for the coding test, as designing a real world wallet system required more throughtout in designing, planing, and discussion with other stakehodler based on business usecases and there will be more risk and auditing system needed to be ready (For example, the cash pool movement, compliance, auditing).

# Notes (Confirmation from email with Seline)

1. Authentication is not required.
2. SGD and JPY currency are supports in wallets for each user
3. Bank Transfer source to deposit
4. Withdrawal (just implement the debit and transaction history.)

# Functional requirements

1. [] Deposit to specify user wallet
2. [] Withdraw from specify user wallet
3. [] Transfer from one user to another user
4. [] Get specify user balance
5. [] Get specify user transaction history

# Non Functional requirements

1. Data consistency is critical as this is a financial system
2. Low latency responses
3. Audit logging for all transactions
4. PII protection (Optional)
5. Minitoring and alerting system (API response time, slow query detector and etc)

# Database design

1. Users entity -> Represent our wallets users
2. Wallets entity -> Each users can have multiple wallets, each wallets must be hold differnt currency, wallet id is a composite key of (username + currency)
3. Transactions entity -> For each transaction make to the wallets such as deposit, withdraw, transfers and etc will be store it in transactions table.
4. Ledgers table -> This is a simple ledger system to store the fund movement of each wallets.

# Assumptions

## Race condition

1. As a user, I should protected from race conditions that could occur during concurrent transactions on my wallet, ensuring my balance remains accurate even when multiple operations happen simultaneously.

## Transaction Log

2. As a user, I should be able to view all the fund movement of my wallets (DEPOSIT, WITHDRAW, TRANSFER).

## Transactions operation for each type

3. For Deposit and Withdraw transactions, each will create One entry into transactions table and ledgers table. For Transfer transaction, 2 entires will be created in ledgers to identify the fund movement of sender and receiver, One entry will be create in Transaction to record the action.

## Avoid transaction stay in unknown status

4. Recognizing that timeouts, network issues, and external system outages are risks to our system, a robust recovery strategy is essential. This strategy must include notifications and the capability for immediate/later transaction retries to minimize disruption.

## Connection

```bash
psql -h localhost -U admin -d wallet
```

Then type "secret" as password
