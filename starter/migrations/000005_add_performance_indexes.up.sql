-- Performance Optimization Indexes

CREATE INDEX IF NOT EXISTS idx_media_type ON media(type);
CREATE INDEX IF NOT EXISTS idx_media_created_at ON media(created_at);
CREATE INDEX IF NOT EXISTS idx_media_url_text ON media USING gin(to_tsvector('english', url));
 
CREATE INDEX IF NOT EXISTS idx_posts_author ON posts(author);
CREATE INDEX IF NOT EXISTS idx_posts_created_at ON posts(created_at);
CREATE INDEX IF NOT EXISTS idx_posts_title_text ON posts USING gin(to_tsvector('english', title));
CREATE INDEX IF NOT EXISTS idx_posts_content_text ON posts USING gin(to_tsvector('english', content));

CREATE INDEX IF NOT EXISTS idx_pages_created_at ON pages(created_at);
CREATE INDEX IF NOT EXISTS idx_pages_title_text ON pages USING gin(to_tsvector('english', title));
CREATE INDEX IF NOT EXISTS idx_pages_content_text ON pages USING gin(to_tsvector('english', content));

CREATE INDEX IF NOT EXISTS idx_post_media_post_id ON post_media(post_id);
CREATE INDEX IF NOT EXISTS idx_post_media_media_id ON post_media(media_id);

CREATE INDEX IF NOT EXISTS idx_posts_author_created_at ON posts(author, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_media_type_created_at ON media(type, created_at DESC);