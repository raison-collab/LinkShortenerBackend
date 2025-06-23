-- Create link_clicks table
CREATE TABLE IF NOT EXISTS link_clicks (
    id BIGSERIAL PRIMARY KEY,
    link_id BIGINT NOT NULL REFERENCES links(id) ON DELETE CASCADE,
    ip_address INET NOT NULL,
    user_agent TEXT,
    referer TEXT,
    country VARCHAR(100),
    city VARCHAR(100),
    clicked_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes
CREATE INDEX idx_link_clicks_link_id ON link_clicks(link_id);
CREATE INDEX idx_link_clicks_clicked_at ON link_clicks(clicked_at);
CREATE INDEX idx_link_clicks_ip_address ON link_clicks(ip_address);
CREATE INDEX idx_link_clicks_country ON link_clicks(country); 