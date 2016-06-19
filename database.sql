CREATE TABLE `pastebin` (
  `id` varchar(30) NOT NULL,
  `hash` char(40) default NULL,
  `data` longtext,
  `delkey` char(40) default NULL,
  PRIMARY KEY (`id`)
);
