-- PostgreSQL Database Schema for Sudanese Podcast Platform

-- Users Table (listeners, podcasters, admins)
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    username VARCHAR(50) UNIQUE NOT NULL,
    password_hash VARCHAR(255),
    full_name VARCHAR(100),
    bio TEXT,
    profile_image_url VARCHAR(255),
    user_type VARCHAR(20) NOT NULL CHECK (user_type IN ('listener', 'podcaster', 'admin')),
    auth_provider VARCHAR(20) CHECK (auth_provider IN ('google', 'apple', 'email', 'guest')),
    auth_provider_id VARCHAR(255),
    is_verified BOOLEAN DEFAULT FALSE,
    preferred_language VARCHAR(10) DEFAULT 'ar-sd',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_login_at TIMESTAMP WITH TIME ZONE,
    UNIQUE (auth_provider, auth_provider_id)
);

-- Podcasts Table
CREATE TABLE podcasts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    podcaster_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    cover_image_url VARCHAR(255),
    rss_url VARCHAR(255) UNIQUE,
    website_url VARCHAR(255),
    language VARCHAR(10) DEFAULT 'ar-sd',
    author VARCHAR(100),
    category VARCHAR(50),
    subcategory VARCHAR(50),
    explicit BOOLEAN DEFAULT FALSE,
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'pending', 'rejected', 'archived')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    last_synced_at TIMESTAMP WITH TIME ZONE
);

-- Episodes Table
CREATE TABLE episodes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    podcast_id UUID NOT NULL REFERENCES podcasts(id) ON DELETE CASCADE,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    audio_url VARCHAR(255) NOT NULL,
    duration INTEGER, -- in seconds
    cover_image_url VARCHAR(255),
    publication_date TIMESTAMP WITH TIME ZONE,
    guid VARCHAR(255), -- from RSS feed
    episode_number INTEGER,
    season_number INTEGER,
    transcript TEXT,
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'pending', 'rejected', 'archived')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (podcast_id, guid)
);

-- Categories Table
CREATE TABLE categories (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(50) NOT NULL,
    description TEXT,
    icon_url VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Podcast-Category Mapping (Many-to-Many)
CREATE TABLE podcast_categories (
    podcast_id UUID REFERENCES podcasts(id) ON DELETE CASCADE,
    category_id UUID REFERENCES categories(id) ON DELETE CASCADE,
    PRIMARY KEY (podcast_id, category_id)
);

-- Subscriptions Table (Listeners to Podcasts)
CREATE TABLE subscriptions (
    listener_id UUID REFERENCES users(id) ON DELETE CASCADE,
    podcast_id UUID REFERENCES podcasts(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (listener_id, podcast_id)
);

-- Playback History
CREATE TABLE playback_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    listener_id UUID REFERENCES users(id) ON DELETE CASCADE,
    episode_id UUID REFERENCES episodes(id) ON DELETE CASCADE,
    position INTEGER NOT NULL, -- in seconds
    completed BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (listener_id, episode_id)
);

-- Downloads Table
CREATE TABLE downloads (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    listener_id UUID REFERENCES users(id) ON DELETE CASCADE,
    episode_id UUID REFERENCES episodes(id) ON DELETE CASCADE,
    status VARCHAR(20) DEFAULT 'completed' CHECK (status IN ('pending', 'completed', 'failed')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (listener_id, episode_id)
);

-- Likes Table
CREATE TABLE likes (
    listener_id UUID REFERENCES users(id) ON DELETE CASCADE,
    episode_id UUID REFERENCES episodes(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (listener_id, episode_id)
);

-- Comments Table
CREATE TABLE comments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    episode_id UUID REFERENCES episodes(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('active', 'hidden', 'flagged')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Playlists Table
CREATE TABLE playlists (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    is_public BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Playlist Items
CREATE TABLE playlist_items (
    playlist_id UUID REFERENCES playlists(id) ON DELETE CASCADE,
    episode_id UUID REFERENCES episodes(id) ON DELETE CASCADE,
    position INTEGER NOT NULL,
    added_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (playlist_id, episode_id)
);

-- Analytics Tables

-- Listen Events Table
CREATE TABLE listen_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    listener_id UUID REFERENCES users(id) ON DELETE SET NULL,
    episode_id UUID REFERENCES episodes(id) ON DELETE CASCADE,
    source VARCHAR(50), -- mobile, web, embed
    started_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    duration INTEGER, -- how long they listened in seconds
    completed BOOLEAN DEFAULT FALSE,
    ip_address VARCHAR(45),
    user_agent TEXT,
    country_code VARCHAR(2),
    city VARCHAR(100)
);

-- Subscription Plans
CREATE TABLE subscription_plans (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    description TEXT,
    price_monthly DECIMAL(10, 2) NOT NULL,
    price_yearly DECIMAL(10, 2) NOT NULL,
    features JSONB NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- User Subscriptions (Premium)
CREATE TABLE user_subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    plan_id UUID NOT NULL REFERENCES subscription_plans(id),
    status VARCHAR(20) NOT NULL CHECK (status IN ('active', 'canceled', 'expired')),
    start_date TIMESTAMP WITH TIME ZONE NOT NULL,
    end_date TIMESTAMP WITH TIME ZONE NOT NULL,
    is_auto_renew BOOLEAN DEFAULT TRUE,
    payment_method VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Payments Table
CREATE TABLE payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    subscription_id UUID REFERENCES user_subscriptions(id),
    amount DECIMAL(10, 2) NOT NULL,
    currency VARCHAR(3) DEFAULT 'SDG',
    payment_method VARCHAR(50) NOT NULL,
    status VARCHAR(20) NOT NULL CHECK (status IN ('pending', 'completed', 'failed', 'refunded')),
    transaction_id VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Advertisements Table
CREATE TABLE advertisements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    advertiser_name VARCHAR(100) NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    audio_url VARCHAR(255) NOT NULL,
    duration INTEGER NOT NULL, -- in seconds
    start_date TIMESTAMP WITH TIME ZONE NOT NULL,
    end_date TIMESTAMP WITH TIME ZONE NOT NULL,
    status VARCHAR(20) DEFAULT 'active' CHECK (status IN ('pending', 'active', 'completed', 'rejected')),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Ad Placements
CREATE TABLE ad_placements (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ad_id UUID NOT NULL REFERENCES advertisements(id) ON DELETE CASCADE,
    podcast_id UUID REFERENCES podcasts(id) ON DELETE CASCADE,
    episode_id UUID REFERENCES episodes(id) ON DELETE CASCADE,
    position VARCHAR(20) NOT NULL CHECK (position IN ('pre-roll', 'mid-roll', 'post-roll')),
    position_seconds INTEGER, -- for mid-roll, seconds into episode
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    CHECK (
        (podcast_id IS NOT NULL AND episode_id IS NULL) OR 
        (podcast_id IS NULL AND episode_id IS NOT NULL)
    )
);

-- Ad Impressions
CREATE TABLE ad_impressions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    ad_placement_id UUID NOT NULL REFERENCES ad_placements(id) ON DELETE CASCADE,
    listener_id UUID REFERENCES users(id) ON DELETE SET NULL,
    episode_id UUID REFERENCES episodes(id) ON DELETE CASCADE,
    listened_duration INTEGER, -- how many seconds of the ad were played
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Indices
CREATE INDEX idx_episodes_podcast_id ON episodes(podcast_id);
CREATE INDEX idx_playback_history_listener_id ON playback_history(listener_id);
CREATE INDEX idx_playback_history_episode_id ON playback_history(episode_id);
CREATE INDEX idx_listen_events_episode_id ON listen_events(episode_id);
CREATE INDEX idx_listen_events_listener_id ON listen_events(listener_id);
CREATE INDEX idx_comments_episode_id ON comments(episode_id);
CREATE INDEX idx_podcasts_status ON podcasts(status);
CREATE INDEX idx_episodes_status ON episodes(status);
CREATE INDEX idx_podcasts_category ON podcasts(category);
CREATE INDEX idx_podcasts_language ON podcasts(language);
CREATE INDEX idx_episodes_publication_date ON episodes(publication_date);
CREATE INDEX idx_user_subscriptions_user_id ON user_subscriptions(user_id);
CREATE INDEX idx_user_subscriptions_status ON user_subscriptions(status);
CREATE INDEX idx_payments_user_id ON payments(user_id);
CREATE INDEX idx_payments_status ON payments(status);