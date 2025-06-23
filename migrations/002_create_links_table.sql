-- Create links table
CREATE TABLE IF NOT EXISTS links (
    id BIGSERIAL PRIMARY KEY,
    short_code VARCHAR(20) UNIQUE NOT NULL,
    original_url TEXT NOT NULL,
    user_id BIGINT REFERENCES users(id) ON DELETE CASCADE,
    clicks BIGINT DEFAULT 0,
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes
CREATE INDEX idx_links_short_code ON links(short_code);
CREATE INDEX idx_links_user_id ON links(user_id);
CREATE INDEX idx_links_is_active ON links(is_active);
CREATE INDEX idx_links_expires_at ON links(expires_at);
CREATE INDEX idx_links_created_at ON links(created_at);

-- Create updated_at trigger
CREATE TRIGGER update_links_updated_at BEFORE UPDATE
    ON links FOR EACH ROW EXECUTE FUNCTION update_updated_at_column(); 