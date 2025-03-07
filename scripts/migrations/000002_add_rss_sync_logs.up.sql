-- scripts/migrations/000002_add_rss_sync_logs.up.sql
CREATE TABLE rss_sync_logs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    podcast_id UUID NOT NULL REFERENCES podcasts(id) ON DELETE CASCADE,
    status VARCHAR(20) NOT NULL CHECK (status IN ('success', 'failure')),
    episodes_added INTEGER DEFAULT 0,
    episodes_updated INTEGER DEFAULT 0,
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_rss_sync_logs_podcast_id ON rss_sync_logs(podcast_id);
CREATE INDEX idx_rss_sync_logs_created_at ON rss_sync_logs(created_at);

-- scripts/migrations/000002_add_rss_sync_logs.down.sql
DROP TABLE IF EXISTS rss_sync_logs;