CREATE TABLE accounts (
    account_id UUID NOT NULL DEFAULT (uuid_generate_v4()),
    first_name VARCHAR(255) NOT NULL,
    last_name VARCHAR(255) NOT NULL,
    PRIMARY KEY (account_id)
);

CREATE TABLE transactions (
    transaction_id UUID NOT NULL DEFAULT (uuid_generate_v4()),
    owner UUID NOT NULL,
    sender UUID NOT NULL,
    receiver UUID NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    amount NUMERIC(9, 2) NOT NULL,
    is_consumed BOOLEAN NOT NULL DEFAULT FALSE,
    PRIMARY KEY (transaction_id),
    FOREIGN KEY (owner) REFERENCES accounts(account_id),
    FOREIGN KEY (sender) REFERENCES accounts(account_id),
    FOREIGN KEY (receiver) REFERENCES accounts(account_id)
    
);