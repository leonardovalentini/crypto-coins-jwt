CREATE DATABASE IF NOT EXISTS auth;

GRANT ALL PRIVILEGES ON auth.* TO 'user'@'%';
FLUSH PRIVILEGES;

USE auth;

-- DROP TABLE IF EXISTS `users`;
CREATE TABLE IF NOT EXISTS users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    username VARCHAR(50) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL
);

-- Insert a test user (password is 'secret123' hashed with bcrypt)
INSERT INTO users (username, password) 
VALUES ('admin', '$2a$10$XMZHoMpCG1o9toxIwCclJeupL11xbilkPikSLFJJmIZ0zntIFGKNm91')
ON DUPLICATE KEY UPDATE id=id;
