CREATE TABLE conversations (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    is_group BOOLEAN NOT NULL DEFAULT FALSE,
    name VARCHAR(255), -- nếu là group
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
