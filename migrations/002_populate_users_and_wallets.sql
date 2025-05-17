-- Insert 10 users
INSERT INTO users (username, password) VALUES
('user1', 'password'),
('user2', 'password'),
('user3', 'password'),
('user4', 'password'),
('user5', 'password'),
('user6', 'password'),
('user7', 'password'),
('user8', 'password'),
('user9', 'password'),
('user10', 'password');

-- Create wallets for each user
INSERT INTO wallets (username, currency, amount, status) VALUES
('user1', 'SGD', 0, 'active'),
('user1', 'JPY', 0, 'active'),
('user2', 'SGD', 0, 'active'),
('user2', 'JPY', 0, 'active'),
('user3', 'SGD', 0, 'active'),
('user3', 'JPY', 0, 'active'),
('user4', 'SGD', 0, 'active'),
('user4', 'JPY', 0, 'active'),
('user5', 'SGD', 0, 'active'),
('user5', 'JPY', 0, 'active'),
('user6', 'SGD', 0, 'active'),
('user6', 'JPY', 0, 'active'),
('user7', 'SGD', 0, 'active'),
('user7', 'JPY', 0, 'active'),
('user8', 'SGD', 0, 'active'),
('user8', 'JPY', 0, 'active'),
('user9', 'SGD', 0, 'active'),
('user9', 'JPY', 0, 'active'),
('user10', 'SGD', 0, 'active'),
('user10', 'JPY', 0, 'active');
