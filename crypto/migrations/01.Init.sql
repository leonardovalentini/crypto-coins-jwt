CREATE DATABASE IF NOT EXISTS crypto;

GRANT ALL PRIVILEGES ON crypto.* TO 'user'@'%';
FLUSH PRIVILEGES;

USE crypto;

-- DROP TABLE IF EXISTS `user_coins`;
CREATE TABLE IF NOT EXISTS `user_coins` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `user_id` varchar(100) NOT NULL,
  `coin` varchar(3) NOT NULL,
  `coin_ref` varchar(3) NOT NULL,
   CONSTRAINT uc_user_coin UNIQUE (user_id, coin, coin_ref),
  PRIMARY KEY (`id`)
) ENGINE=InnoDB CHARSET=latin1;
INSERT INTO `user_coins` VALUES
	(1,'1','BTC','USD'),
	(2,'1','ETH','USD');


-- DROP TABLE IF EXISTS `prices`;
CREATE TABLE IF NOT EXISTS `prices` (
  `id` int(11) NOT NULL AUTO_INCREMENT,
  `coin` varchar(3) NOT NULL,
  `coin_ref` varchar(3) NOT NULL,
  `date` DATETIME NOT NULL,
  `price` DECIMAL(24, 12) NOT NULL,
  `site` VARCHAR(255) NOT NULL,
  PRIMARY KEY (`id`)
) ENGINE=InnoDB CHARSET=latin1;