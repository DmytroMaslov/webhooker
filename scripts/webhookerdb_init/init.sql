-- Create Orders table
CREATE TABLE Orders (
    OrderID VARCHAR(37) PRIMARY KEY,
    UserId VARCHAR(37) NOT NULL,
    OrderStatus VARCHAR(50) NOT NULL,
    IsFinal BOOLEAN NOT NULL,
    CreateAt TIMESTAMP NOT NULL,
    UpdateAt TIMESTAMP NOT NULL
);
-- Create Events table
CREATE TABLE JustPayEvents (
    EventID VARCHAR(37) PRIMARY KEY NOT NULL,
    OrderID VARCHAR(37) NOT NULL,
    UserID VARCHAR(37) NOT NULL,
    OrderStatus VARCHAR(50) NOT NULL,
    CreateAt TIMESTAMP NOT NULL,
    UpdateAt TIMESTAMP NOT NULL
);

-- create index
CREATE INDEX just_pay_events_orderid ON JustPayEvents (OrderID);

-- insert test orders
INSERT INTO Orders (OrderID, UserID, OrderStatus, IsFinal, CreateAt, UpdateAt)
VALUES('97a96c29-7631-4cbc-9559-f8866fb03392', '2c127d70-3b9b-4743-9c2e-74b9f617029f', 'cool_order_created', false, '2022-10-10 11:30:30', '2022-10-10 11:30:30');

INSERT INTO Orders (OrderID, UserID, OrderStatus, IsFinal, CreateAt, UpdateAt)
VALUES('97a96c29-7631-4cbc-9559-f8866fb03393', '3c127d70-3b9b-4743-9c2e-74b9f617029f', 'sbu_verification_pending', false, '2022-10-10 11:30:31', '2022-10-10 11:30:31');

INSERT INTO Orders (OrderID, UserID, OrderStatus, IsFinal, CreateAt, UpdateAt)
VALUES('97a96c29-7631-4cbc-9559-f8866fb03394', '4c127d70-3b9b-4743-9c2e-74b9f617029f', 'chinazes', true, '2022-10-10 11:30:30', '2022-10-10 11:30:33');

