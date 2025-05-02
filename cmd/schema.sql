-- 1. ENUM type for service status
CREATE TYPE status AS ENUM (
  'created',
  'ongoing',
  'done',
  'cancelled'
);

-- 2. Users table
CREATE TABLE IF NOT EXISTS users (
  id SERIAL PRIMARY KEY,
  email VARCHAR,
  password VARCHAR
);

-- 3. Profiles table: 1-to-1 with users
CREATE TABLE IF NOT EXISTS profiles (
  id SERIAL PRIMARY KEY,
  user_id INTEGER NOT NULL UNIQUE,
  username VARCHAR NOT NULL,
  tokens INTEGER NOT NULL,
  rating DOUBLE PRECISION NOT NULL,
  joined_at TIMESTAMP NOT NULL,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- 4. Posts table: many posts per user
CREATE TABLE IF NOT EXISTS posts (
  id SERIAL PRIMARY KEY,
  user_id INTEGER NOT NULL,
  title TEXT,
  description TEXT,
  created_at TIMESTAMP,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- 5. Services table: a user offers a service tied to a post
CREATE TABLE IF NOT EXISTS services (
  id SERIAL PRIMARY KEY,
  user_id INTEGER NOT NULL,
  post_id INTEGER NOT NULL,
  cost INTEGER,
  FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
  FOREIGN KEY (post_id) REFERENCES posts(id) ON DELETE CASCADE
);

-- 6. Transactions table: a provider serves a client
CREATE TABLE IF NOT EXISTS transactions (
  id SERIAL PRIMARY KEY,
  provider_id INTEGER NOT NULL,
  client_id INTEGER NOT NULL,
  service_status status NOT NULL,
  FOREIGN KEY (provider_id) REFERENCES users(id) ON DELETE CASCADE,
  FOREIGN KEY (client_id) REFERENCES users(id) ON DELETE CASCADE
);