CREATE TABLE IF NOT EXISTS tickets (
    id INTEGER PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT NOT NULL,
    requester_email TEXT NOT NULL DEFAULT '',
    status TEXT NOT NULL DEFAULT 'open',
    created_at DATETIME NOT NULL
);

CREATE TABLE IF NOT EXISTS kb_articles (
    id INTEGER PRIMARY KEY,
    title TEXT NOT NULL,
    body TEXT NOT NULL,
    tags TEXT NOT NULL DEFAULT ''
);

CREATE VIRTUAL TABLE IF NOT EXISTS kb_fts USING fts5(
    title, body, tags,
    content='kb_articles', content_rowid='id',
    tokenize='porter unicode61'
);

CREATE TRIGGER IF NOT EXISTS kb_articles_ai 
AFTER INSERT ON kb_articles BEGIN
    INSERT INTO kb_fts(rowid, title, body, tags) VALUES (new.id, new.title, new.body, new.tags);
END;
CREATE TRIGGER IF NOT EXISTS kb_articles_ad
AFTER DELETE ON kb_articles BEGIN
    INSERT INTO kb_fts(kb_fts, rowid, title, body, tags) VALUES ('delete', old.id, old.title, old.body, old.tags);
END;
CREATE TRIGGER IF NOT EXISTS kb_articles_au
AFTER UPDATE ON kb_articles BEGIN
    INSERT INTO kb_fts(kb_fts, rowid, title, body, tags) VALUES ('delete', old.id, old.title, old.body, old.tags);
    INSERT INTO kb_fts(rowid, title, body, tags) VALUES (new.id, new.title, new.body, new.tags);
END;

CREATE TABLE IF NOT EXISTS triages (
    id INTEGER PRIMARY KEY,
    ticket_id INTEGER NOT NULL REFERENCES tickets(id),
    priority TEXT NOT NULL,
    category TEXT NOT NULL,
    summary TEXT NOT NULL,
    steps TEXT NOT NULL,
    article_ids TEXT NOT NULL,
    created_at DATETIME NOT NULL
);