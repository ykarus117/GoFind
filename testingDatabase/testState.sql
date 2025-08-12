INSERT INTO Users (username,pwd,salt) VALUES
    ('user0', x'2BD10B3E6CACBB1302D9AAB7AEC6E67494917A85AC18B99E4F514956FC5B252B',x'7E590D74C6B552F53856ED5E2B477323');

INSERT INTO Items (name, tags, owned_by) VALUES
                                             ( 'Item One', 'tag1,tag2', 'obj1'),
                                             ( 'Item Two', 'tag3', 'obj2'),
                                             ('Item Four', 'tag5', 'obj2'),
                                             ( 'Item Seven', 'tag1' , 'obj5'),
                                             ('Item Three', 'tag4', 'obj3');

INSERT INTO Objects (id, name, type, owner_id, ownedBy_user) VALUES
                                                                 ('obj3', 'Object Three', 'Type C', NULL, 10),
                                                                 ('obj1', 'Object One', 'Type A',NULL, 10),
                                                                 ('obj2', 'Object Two', 'Type D', 'obj1', 10),
                                                                 ('obj5', 'Object five', 'Type C', 'obj2', 10),
                                                                 ('obj4', 'Object Four', 'Type A', NUll, 10),
                                                                 ('obj6', 'Object Six', 'Type A', NULL,10),
                                                                 ('obj7', 'Object Seven', 'Type A', 'obj6', 10),
                                                                 ('obj9', 'Object Nine', 'Type D', NULL, 2);