-- migrations/000002_create_transcripts_table.down.sql
DROP INDEX IF EXISTS idx_transcripts_video_created;
DROP INDEX IF EXISTS idx_transcripts_segments;
DROP INDEX IF EXISTS idx_transcripts_raw_text_fts;
DROP INDEX IF EXISTS idx_transcripts_created_at;
DROP INDEX IF EXISTS idx_transcripts_source;
DROP INDEX IF EXISTS idx_transcripts_language;
DROP INDEX IF EXISTS idx_transcripts_video_id;
DROP TABLE IF EXISTS transcripts;