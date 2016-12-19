CREATE TABLE `pastebin` (
  `id` varchar(30) NOT NULL,
  `title` varchar(50) default NULL,
  `hash` char(40) default NULL,
  `data` longtext,
  `delkey` char(40) default NULL,
  `expiry` int,
  PRIMARY KEY (`id`)
);
