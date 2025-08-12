CREATE TABLE Objects (
                         id string NOT NULL PRIMARY KEY,
                         name string NOT NULL,
                         type string NOT NULL,
                         created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
                         updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
                         description string,
                         owner_id string,
                         ownedBy_user INTEGER,
                         FOREIGN KEY (owner_id) REFERENCES Objects (id) DEFERRABLE,
                         FOREIGN KEY (ownedBy_user) REFERENCES Users (id) ON DELETE CASCADE
);

CREATE TABLE Items (
                       id INTEGER PRIMARY KEY,
                       name string NOT NULL,
                       tags string,
                       created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
                       updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
                       quantity INTEGER NOT NULL DEFAULT 1,
                       owned_by string NOT NULL,
                       ownedBy_user INTEGER,
                       FOREIGN KEY (owned_by) REFERENCES Objects (id) DEFERRABLE,
                       FOREIGN KEY (ownedBy_user) REFERENCES Users (id) ON DELETE CASCADE
);

CREATE TABLE Users (
                       id INTEGER PRIMARY KEY,
                       username TEXT NOT NULL,
                       pwd BLOB NOT NULL,
                       salt BLOB NOT NULL
);

CREATE INDEX idx_objects_owner ON Objects (owner_id);
CREATE INDEX idx_items_owned_by ON Items (owned_by);
CREATE INDEX idx_items_tags ON Items (tags);
CREATE INDEX idx_objects_id ON Objects (id);
CREATE INDEX idx_items_name ON Items (name);
CREATE INDEX idx_usernames ON Users (username);
CREATE INDEX idx_ownedBy_user ON Objects (ownedBy_user);