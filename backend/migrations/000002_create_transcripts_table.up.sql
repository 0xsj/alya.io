CREATE TABLE IF NOT EXISTS transcripts (
    id VARCHAR(36) PRIMARY KEY,
    video_id VARCHAR(36) NOT NULL,
    language VARCHAR(10) NOT NULL,
    segments JSONB NOT NULL,
    raw_text TEXT NOT NULL,
    source VARCHAR(50) NOT NULL,
    processed_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL,
    
    -- Foreign key constraint to videos table
    CONSTRAINT fk_transcripts_video_id 
        FOREIGN KEY (video_id) 
        REFERENCES videos(id) 
        ON DELETE CASCADE
);

-- Create indexes for common queries
CREATE INDEX idx_transcripts_video_id ON transcripts(video_id);
CREATE INDEX idx_transcripts_language ON transcripts(language);
CREATE INDEX idx_transcripts_source ON transcripts(source);
CREATE INDEX idx_transcripts_created_at ON transcripts(created_at);

-- Full text search index for raw_text
CREATE INDEX idx_transcripts_raw_text_fts ON transcripts USING GIN(to_tsvector('english', raw_text));

-- GIN index for JSONB segments for advanced querying
CREATE INDEX idx_transcripts_segments ON transcripts USING GIN(segments);

-- Composite index for video_id + created_at (for getting latest transcript)
CREATE INDEX idx_transcripts_video_created ON transcripts(video_id, created_at DESC);

-- Add constraints
ALTER TABLE transcripts ADD CONSTRAINT check_language_format 
    CHECK (language ~ '^[a-z]{2}(-[A-Z]{2})?$');
    
ALTER TABLE transcripts ADD CONSTRAINT check_source_not_empty 
    CHECK (source != '');
    
ALTER TABLE transcripts ADD CONSTRAINT check_raw_text_not_empty 
    CHECK (raw_text != '');