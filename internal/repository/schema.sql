PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS users(
    id INTEGER NOT NULL PRIMARY KEY,
    password VARCHAR(72) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL
);

CREATE TABLE IF NOT EXISTS profiles(
    id INTEGER NOT NULL PRIMARY KEY,
    user_id INTEGER NOT NULL UNIQUE,
    username VARCHAR(50) NOT NULL,
    tokens INTEGER NOT NULL,
    date_joined DATE NOT NULL,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS sessions (
	token TEXT PRIMARY KEY,
	data BLOB NOT NULL,
	expiry REAL NOT NULL
);

CREATE INDEX sessions_expiry_idx ON sessions(expiry);