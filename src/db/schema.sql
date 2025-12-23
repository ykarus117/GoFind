CREATE TABLE Objects (
                         id INTEGER PRIMARY KEY,
                         name string NOT NULL,
                         created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
                         updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
                         description string DEFAULT '',
                         owner_id INTEGER DEFAULT NULL,
                         ownedBy_user INTEGER NOT NULL,
                         FOREIGN KEY (owner_id) REFERENCES Objects (id) ON UPDATE CASCADE ON DELETE SET NULL DEFERRABLE,
                         FOREIGN KEY (ownedBy_user) REFERENCES Users (id) ON DELETE CASCADE ON UPDATE CASCADE
);

CREATE TABLE Items (
                       id INTEGER PRIMARY KEY,
                       name string NOT NULL,
                       description string DEFAULT '',
                       created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
                       updated_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,
                       quantity INTEGER NOT NULL DEFAULT 1,
                       owned_by INTEGER DEFAULT NULL,
                       ownedBy_user INTEGER NOT NULL,
                       FOREIGN KEY (owned_by) REFERENCES Objects (id) ON UPDATE CASCADE ON DELETE SET NULL DEFERRABLE,
                       FOREIGN KEY (ownedBy_user) REFERENCES Users (id) ON UPDATE CASCADE ON DELETE CASCADE
);

CREATE TABLE Users (
                       id INTEGER PRIMARY KEY,
                       username TEXT NOT NULL,
                       pwd BLOB NOT NULL,
                       salt BLOB NOT NULL
);

CREATE TABLE  tags_items (
                             item_id INTEGER,
                             tag_id INTEGER,
                             FOREIGN KEY (item_id) REFERENCES Items(id) ON DELETE CASCADE DEFERRABLE,
                             FOREIGN KEY (tag_id) REFERENCES Tags(id) ON DELETE CASCADE DEFERRABLE
);

CREATE TABLE tags_objects (
    object_id INTEGER,
    tag_id INTEGER,
    FOREIGN KEY (object_id) REFERENCES Objects(id) ON DELETE CASCADE ON UPDATE CASCADE,
    FOREIGN KEY (tag_id) REFERENCES  Tags(id) ON DELETE CASCADE
);

CREATE TABLE Tags (
    id INTEGER PRIMARY KEY, name STRING UNIQUE NOT NULL
);

CREATE INDEX idx_objects_owner ON Objects (owner_id);
CREATE INDEX idx_items_owned_by ON Items (owned_by);
CREATE INDEX idx_items_name ON Items (name);
CREATE INDEX idx_usernames ON Users (username);
CREATE INDEX idx_ownedBy_user ON Objects (ownedBy_user);

CREATE UNIQUE INDEX named_items_users ON Objects (name, ownedBy_user);
CREATE UNIQUE INDEX tags_items_index on tags_items (item_id, tag_id);
CREATE UNIQUE INDEX tags_obj_index on tags_objects (object_id, tag_id);
CREATE UNIQUE INDEX nesting_objects on Objects (name, owner_id, ownedBy_user);


CREATE VIEW 'user0_objects' AS SELECT * FROM Objects WHERE ownedBy_user = 0;
CREATE VIEW 'user0_items' AS SELECT * FROM Items WHERE ownedBy_user = 0;