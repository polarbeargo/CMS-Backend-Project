-- Drop performance indexes
DROP INDEX IF EXISTS idx_media_type;
DROP INDEX IF EXISTS idx_media_created_at;
DROP INDEX IF EXISTS idx_media_url_text;

DROP INDEX IF EXISTS idx_posts_author;
DROP INDEX IF EXISTS idx_posts_created_at;
DROP INDEX IF EXISTS idx_posts_title_text;
DROP INDEX IF EXISTS idx_posts_content_text;

DROP INDEX IF EXISTS idx_pages_created_at;
DROP INDEX IF EXISTS idx_pages_title_text;
DROP INDEX IF EXISTS idx_pages_content_text;

DROP INDEX IF EXISTS idx_post_media_post_id;
DROP INDEX IF EXISTS idx_post_media_media_id;

DROP INDEX IF EXISTS idx_posts_author_created_at;
DROP INDEX IF EXISTS idx_media_type_created_at;