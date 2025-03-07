-- scripts/seed/seed_data.sql

-- Insert sample categories
INSERT INTO categories (id, name, description, icon_url) VALUES
    ('1f8a5d74-6d12-4c32-a963-48190907c0ec', 'Politics', 'Political discussions and news', '/icons/politics.png'),
    ('550e8400-e29b-41d4-a716-446655440000', 'Culture', 'Cultural topics and discussions', '/icons/culture.png'),
    ('7ae1cc5c-5a88-4d5d-b371-64b5f9a3c9a3', 'Music', 'Music reviews and discussions', '/icons/music.png'),
    ('e0561b9c-7167-4588-a0b5-2bc576258cde', 'Technology', 'Technology news and discussions', '/icons/technology.png'),
    ('c80dc35c-1e6f-4ee2-8dd1-31d0f53698be', 'History', 'Historical topics and discussions', '/icons/history.png'),
    ('d92e8994-f5a8-4c5f-a379-3a6e5cab4e16', 'Education', 'Educational content and discussions', '/icons/education.png');

-- Insert admin user
INSERT INTO users (id, email, username, password_hash, full_name, user_type, auth_provider, is_verified, preferred_language) VALUES
    ('6d2db97c-a18b-4a8e-8674-62c5a0621b6c', 'admin@example.com', 'admin', '$2a$10$XlUl5K6UzVEkF6tRywOQK.PaYwYGCt4jvZelgYiskKjS.Nq7S7xrG', 'Admin User', 'admin', 'email', true, 'ar-sd');

-- Insert sample podcasters
INSERT INTO users (id, email, username, password_hash, full_name, bio, user_type, auth_provider, is_verified, preferred_language) VALUES
    ('7c83061e-8d76-46c9-a353-53d6be1220ea', 'podcaster1@example.com', 'podcaster1', '$2a$10$XlUl5K6UzVEkF6tRywOQK.PaYwYGCt4jvZelgYiskKjS.Nq7S7xrG', 'Podcaster One', 'I create podcasts about Sudanese culture', 'podcaster', 'email', true, 'ar-sd'),
    ('bb7b5a1a-1a84-4102-b53e-22e5f0e7b43f', 'podcaster2@example.com', 'podcaster2', '$2a$10$XlUl5K6UzVEkF6tRywOQK.PaYwYGCt4jvZelgYiskKjS.Nq7S7xrG', 'Podcaster Two', 'I discuss Sudanese politics and history', 'podcaster', 'email', true, 'ar-sd');

-- Insert sample listeners
INSERT INTO users (id, email, username, password_hash, full_name, user_type, auth_provider, is_verified, preferred_language) VALUES
    ('a31d67c6-e85a-4c73-b0ab-a4a290102393', 'listener1@example.com', 'listener1', '$2a$10$XlUl5K6UzVEkF6tRywOQK.PaYwYGCt4jvZelgYiskKjS.Nq7S7xrG', 'Listener One', 'listener', 'email', true, 'ar-sd'),
    ('e8c72153-1f3c-4b0a-8a3d-89e0c7f5cc0a', 'listener2@example.com', 'listener2', '$2a$10$XlUl5K6UzVEkF6tRywOQK.PaYwYGCt4jvZelgYiskKjS.Nq7S7xrG', 'Listener Two', 'listener', 'email', true, 'ar-sd');

-- Insert sample podcasts
INSERT INTO podcasts (id, podcaster_id, title, description, cover_image_url, rss_url, language, author, category, status) VALUES
    ('83aec0d2-6c9b-4872-9e80-a6a5f3202b91', '7c83061e-8d76-46c9-a353-53d6be1220ea', 'Sudanese Culture Today', 'A podcast exploring the rich cultural heritage of Sudan', '/images/podcast1.jpg', 'https://example.com/rss1', 'ar-sd', 'Podcaster One', 'Culture', 'active'),
    ('bac18a18-7c9a-4a17-9897-64c6518f808f', 'bb7b5a1a-1a84-4102-b53e-22e5f0e7b43f', 'Political Discourse', 'Discussions about Sudanese politics and governance', '/images/podcast2.jpg', 'https://example.com/rss2', 'ar-sd', 'Podcaster Two', 'Politics', 'active');

-- Associate podcasts with categories
INSERT INTO podcast_categories (podcast_id, category_id) VALUES
    ('83aec0d2-6c9b-4872-9e80-a6a5f3202b91', '550e8400-e29b-41d4-a716-446655440000'),
    ('bac18a18-7c9a-4a17-9897-64c6518f808f', '1f8a5d74-6d12-4c32-a963-48190907c0ec');

-- Insert sample episodes
INSERT INTO episodes (id, podcast_id, title, description, audio_url, duration, publication_date, guid, status) VALUES
    ('f4b8a17e-8c5f-4d78-9a19-39e33c83d46a', '83aec0d2-6c9b-4872-9e80-a6a5f3202b91', 'Traditional Music of Sudan', 'Exploring the diverse musical traditions across Sudan', '/audio/episode1.mp3', 1800, NOW() - INTERVAL '7 days', 'guid-episode1', 'active'),
    ('2c7d5d65-6a5c-44a0-9c36-1e4297198a7d', '83aec0d2-6c9b-4872-9e80-a6a5f3202b91', 'Sudanese Cuisine', 'Delving into the delicious world of Sudanese food', '/audio/episode2.mp3', 2100, NOW() - INTERVAL '3 days', 'guid-episode2', 'active'),
    ('7de1b5c9-8b9a-4d2a-9e15-4e8a9a7b6a65', 'bac18a18-7c9a-4a17-9897-64c6518f808f', 'Current Political Climate', 'Analyzing the current political situation in Sudan', '/audio/episode3.mp3', 2400, NOW() - INTERVAL '5 days', 'guid-episode3', 'active');

-- Add subscriptions
INSERT INTO subscriptions (listener_id, podcast_id) VALUES
    ('a31d67c6-e85a-4c73-b0ab-a4a290102393', '83aec0d2-6c9b-4872-9e80-a6a5f3202b91'),
    ('a31d67c6-e85a-4c73-b0ab-a4a290102393', 'bac18a18-7c9a-4a17-9897-64c6518f808f'),
    ('e8c72153-1f3c-4b0a-8a3d-89e0c7f5cc0a', 'bac18a18-7c9a-4a17-9897-64c6518f808f');

-- Add playback history
INSERT INTO playback_history (id, listener_id, episode_id, position, completed) VALUES
    ('c5d8a5a7-6c9b-4f1e-9d3a-7b8c9a1d2e3f', 'a31d67c6-e85a-4c73-b0ab-a4a290102393', 'f4b8a17e-8c5f-4d78-9a19-39e33c83d46a', 1200, false),
    ('a9b8c7d6-e5f4-3g2h-1i0j-k9l8m7n6o5p', 'a31d67c6-e85a-4c73-b0ab-a4a290102393', '2c7d5d65-6a5c-44a0-9c36-1e4297198a7d', 2100, true),
    ('f1e2d3c4-b5a6-7890-1234-567890abcdef', 'e8c72153-1f3c-4b0a-8a3d-89e0c7f5cc0a', '7de1b5c9-8b9a-4d2a-9e15-4e8a9a7b6a65', 1500, false);

-- Add likes
INSERT INTO likes (listener_id, episode_id) VALUES
    ('a31d67c6-e85a-4c73-b0ab-a4a290102393', 'f4b8a17e-8c5f-4d78-9a19-39e33c83d46a'),
    ('a31d67c6-e85a-4c73-b0ab-a4a290102393', '2c7d5d65-6a5c-44a0-9c36-1e4297198a7d'),
    ('e8c72153-1f3c-4b0a-8a3d-89e0c7f5cc0a', '7de1b5c9-8b9a-4d2a-9e15-4e8a9a7b6a65');

-- Add comments
INSERT INTO comments (id, user_id, episode_id, content) VALUES
    ('a1b2c3d4-e5f6-g7h8-i9j0-k1l2m3n4o5p6', 'a31d67c6-e85a-4c73-b0ab-a4a290102393', 'f4b8a17e-8c5f-4d78-9a19-39e33c83d46a', 'Fantastic episode! I learned so much about Sudanese music.'),
    ('b2c3d4e5-f6g7-h8i9-j0k1-l2m3n4o5p6q7', 'e8c72153-1f3c-4b0a-8a3d-89e0c7f5cc0a', '7de1b5c9-8b9a-4d2a-9e15-4e8a9a7b6a65', 'Great analysis of the current situation. Looking forward to more!');

-- NOTE: The password hash used above ('$2a$10$XlUl5K6UzVEkF6tRywOQK.PaYwYGCt4jvZelgYiskKjS.Nq7S7xrG') 
-- corresponds to the password 'password123' - this is for testing purposes only