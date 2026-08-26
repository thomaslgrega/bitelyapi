-- Serves the Feed's ordering (ADR-0005); the shape matches its ORDER BY exactly.
CREATE INDEX recipes_feed_idx
  ON recipes (created_at DESC NULLS LAST, id DESC);
