CREATE TABLE IF NOT EXISTS repositories (
    id TEXT PRIMARY KEY,
    repo_id INTEGER UNIQUE NOT NULL,
    name TEXT NOT NULL,
    owner TEXT NOT NULL,
    description TEXT,
    url TEXT NOT NULL,
    language TEXT,
    forks_count INTEGER NOT NULL,
    stars_count INTEGER NOT NULL,
    open_issues INTEGER NOT NULL,
    watchers_count INTEGER NOT NULL,
    is_synced_to_start_time INTEGER NOT NULL DEFAULT 0, -- 0 = false, 1 = true
    is_active INTEGER NOT NULL DEFAULT 1, -- 0 = false, 1 = true
    commit_tracking_start_time TIMESTAMP,
    last_fetched_at TIMESTAMP,
    last_fetched_commit_time TIMESTAMP,
    repo_created_at TIMESTAMP NOT NULL,
    repo_updated_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_repositories_name ON repositories (name);
CREATE INDEX IF NOT EXISTS idx_repositories_owner ON repositories (owner);


CREATE TABLE IF NOT EXISTS commits (
    id TEXT PRIMARY KEY,
    sha TEXT UNIQUE NOT NULL,
    repository_id TEXT NOT NULL REFERENCES repositories(id) ON DELETE CASCADE,
    repo_name TEXT NOT NULL,
    repo_owner TEXT NOT NULL,
    message TEXT NOT NULL,
    author TEXT NOT NULL,
    author_email TEXT NOT NULL,
    date TIMESTAMP NOT NULL,
    url TEXT NOT NULL,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_commits_repo_name ON commits (repo_name);
CREATE INDEX IF NOT EXISTS idx_repo_owner ON commits (repo_owner);
CREATE INDEX IF NOT EXISTS idx_commits_date ON commits (date);


CREATE TABLE IF NOT EXISTS tasks (
    id TEXT PRIMARY KEY,
    repository_id TEXT NOT NULL REFERENCES repositories(id) ON DELETE CASCADE,
    repo_name TEXT NOT NULL,
    repo_owner TEXT NOT NULL,
    status TEXT NOT NULL,
    completed_at TIMESTAMP,
    error_message TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP
);
