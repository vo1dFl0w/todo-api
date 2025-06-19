CREATE TABLE tasks (
    user_id BIGINT NOT NULL,
    task_id BIGSERIAL PRIMARY KEY,
    title VARCHAR NOT NULL,
    description VARCHAR,
    deadline TIMESTAMP,
    complete BOOLEAN DEFAULT false,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);