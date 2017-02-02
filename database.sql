CREATE TABLE `pastebin` (
  `id` varchar(30) NOT NULL,
  `title` varchar(50) default NULL,
  `hash` char(40) default NULL,
  `data` longtext,
  `delkey` char(40) default NULL,
  `expiry` int,
  `userid` int,
  PRIMARY KEY (`id`)
);

CREATE TABLE `accounts` (
  `id` varchar(30) NOT NULL,
  `email` varchar(255) NOT NULL,
  `pass` varchar(255) NOT NULL,
  PRIMARY KEY (`id`)
);
