CREATE TABLE `pastebin` (
  `id` varchar(30) NOT NULL,
  `title` varchar(50) default NULL,
  `hash` char(40) default NULL,
  `data` longtext,
  `delkey` char(40) default NULL,
  `expiry` int,
  `userid` varchar(255),
  PRIMARY KEY (`id`)
);

CREATE TABLE `accounts` (
  `email` varchar(255) NOT NULL,
  `password` varchar(255) NOT NULL,
  `key` varchar(255) NOT NULL,
  PRIMARY KEY (`key`)
);
