CREATE TABLE `pastebin` (
  `id` varchar(30) NOT NULL,
  `title` char(20) default NULL,
  `hash` char(40) default NULL,
  `data` longtext,
  `delkey` char(40) default NULL,
  `expiry` TIMESTAMP,
  PRIMARY KEY (`id`)
);
