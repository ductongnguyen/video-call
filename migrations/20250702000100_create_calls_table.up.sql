-- Tạo enum call_status
CREATE TYPE call_status AS ENUM ('initiated', 'ringing', 'active', 'ended', 'rejected', 'missed');

-- Tạo bảng calls
CREATE TABLE calls (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    caller_id UUID NOT NULL REFERENCES users(id) ON DELETE SET NULL,
    callee_id UUID NOT NULL REFERENCES users(id) ON DELETE SET NULL,
    initiated_id UUID NOT NULL REFERENCES users(id) ON DELETE SET NULL,
    status call_status NOT NULL,
    initiated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    answered_at TIMESTAMPTZ,
    ended_at TIMESTAMPTZ,
    CONSTRAINT check_caller_ne_callee CHECK (caller_id <> callee_id)
);

-- Indexes để tăng tốc độ truy vấn
CREATE INDEX idx_calls_caller_id ON calls(caller_id);
CREATE INDEX idx_calls_callee_id ON calls(callee_id);
CREATE INDEX idx_calls_initiated_id  ON calls(initiated_id );
CREATE INDEX idx_calls_status ON calls(status); 