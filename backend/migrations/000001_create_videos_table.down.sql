-- migrations/000001_create_videos_table.down.sql
DROP TRIGGER IF EXISTS tsvector_update ON videos;
DROP FUNCTION IF EXISTS videos_tsvector_update_trigger();
DROP TABLE IF EXISTS videos;