INSERT INTO Users (id, username,pwd,salt) VALUES
    (0,'user0', x'2BD10B3E6CACBB1302D9AAB7AEC6E67494917A85AC18B99E4F514956FC5B252B',x'7E590D74C6B552F53856ED5E2B477323'),
    (1, 'user1', 'test','test2');

INSERT INTO Objects (id, name, description ,owner_id, ownedBy_user) VALUES
                                                                 (0,'obj3',  'Im object 3!',NULL ,0),
                                                                 (1,'obj1',  'Im object 1!' ,NULL, 0),
                                                                 (2,'obj2',  'Im object 2!', 1, 0),
                                                                 (3,'obj5',   'Im object 5!',2, 0),
                                                                 (4,'obj4',  'Im object 4!', NULL, 0),
                                                                 (5,'obj6',  'Im object 6!',NULL,0),
                                                                 (6,'obj7',   'Im object 7!',5, 0),
                                                                 (7,'obj9',  'Im object 9!',NULL, 1);

INSERT INTO Items (id, name, owned_by, ownedBy_user) VALUES
                                                         (0,'Item One',1, 0),
                                                         (1, 'Item Two',  2, 0),
                                                         (2,'Item Four', 2,0),
                                                         ( 3,'Item Seven' , 3,0),
                                                         (4,'Item Three', 0,0),
                                                         (5, 'Item Five', NULL, 0);

INSERT INTO Tags (id,name) VALUES
                            (0,'tag1'), (1,'tag2'),(2,'tag3'), (3,'tag4');

INSERT INTO tags_items (item_id, tag_id) VALUES
                                             (0, 1),
                                             (0, 2),
                                             (1,3),
                                             (3, 0);