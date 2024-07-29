DROP TABLE IF EXISTS Person;
DROP TABLE IF EXISTS Appl;

CREATE TABLE Person (
    id SERIAL PRIMARY KEY,
    token TEXT UNIQUE,
    login TEXT UNIQUE NOT NULL,
    password TEXT UNIQUE NOT NULL,
    full_name TEXT NOT NULL,
    address TEXT NOT NULL,
    passport_series_number TEXT NOT NULL,
    passport_issue_date TEXT NOT NULL,
    passport_issue_code TEXT NOT NULL,
    passport_issue_authority TEXT NOT NULL,
    consent_text TEXT NOT NULL,
    last_name TEXT,
    first_name TEXT,
    middle_name TEXT,
	birth_date TEXT,
	gender TEXT,
	email TEXT,
	additional_email TEXT,
	phone TEXT,
	mobile TEXT,
	inn TEXT,
	snils TEXT,
	company_name TEXT,
	short_company_name TEXT,
	ogrn TEXT,
	inn2 TEXT,
	kpp TEXT
);

CREATE TABLE Appl (
    login TEXT,
    number TEXT,
    project TEXT,
    amount TEXT,
    selection TEXT,
    status TEXT
);

