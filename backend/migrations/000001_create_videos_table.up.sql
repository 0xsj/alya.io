-- migrations/000001_create_videos_table.up.sql
CREATE TABLE IF NOT EXISTS videos (
    id VARCHAR(36) PRIMARY KEY,
    youtube_id VARCHAR(20) NOT NULL UNIQUE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    url VARCHAR(255) NOT NULL,
    thumbnail_url VARCHAR(255),
    status VARCHAR(20) NOT NULL,
    visibility VARCHAR(20) NOT NULL,
    duration BIGINT,
    language VARCHAR(10),
    transcript_id VARCHAR(36),
    summary_id VARCHAR(36),
    tags TEXT[],
    channel VARCHAR(255),
    channel_id VARCHAR(50),
    views BIGINT,
    like_count BIGINT,
    comment_count BIGINT,
    published_at TIMESTAMP WITH TIME ZONE,
    processed_at TIMESTAMP WITH TIME ZONE,
    error_message TEXT,
    created_by VARCHAR(36) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL,
    
    CHECK (status IN ('pending', 'processing', 'completed', 'failed')),
    CHECK (visibility IN ('public', 'private'))
);

-- Create indexes for common queries
CREATE INDEX idx_videos_youtube_id ON videos(youtube_id);
CREATE INDEX idx_videos_status ON videos(status);
CREATE INDEX idx_videos_created_by ON videos(created_by);
CREATE INDEX idx_videos_created_at ON videos(created_at);