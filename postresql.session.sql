CREATE TABLE user_profile (
	id_account INTEGER PRIMARY KEY REFERENCES account(id_account),
	first_name VARCHAR(50),
	last_name VARCHAR(50),
	email VARCHAR(100) UNIQUE,
	subscription_status VARCHAR(20) DEFAULT 'inactive',
	subscription_expiry TIMESTAMP
);