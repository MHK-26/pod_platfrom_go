-- Drop tables in reverse order of creation to handle dependencies
DROP TABLE IF EXISTS ad_impressions;
DROP TABLE IF EXISTS ad_placements;
DROP TABLE IF EXISTS advertisements;
DROP TABLE IF EXISTS listen_events;
DROP TABLE IF EXISTS playlist_items;
DROP TABLE IF EXISTS playlists;
DROP TABLE IF EXISTS comments;
DROP TABLE IF EXISTS likes;
DROP TABLE IF EXISTS downloads;
DROP TABLE IF EXISTS playback_history;
DROP TABLE IF EXISTS subscriptions;
DROP TABLE IF EXISTS podcast_categories;
DROP TABLE IF EXISTS categories;
DROP TABLE IF EXISTS episodes;
DROP TABLE IF EXISTS podcasts;
DROP TABLE IF EXISTS users;