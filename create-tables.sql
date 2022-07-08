DROP TABLE IF EXISTS pages;
CREATE TABLE pages (
  PageID     INT AUTO_INCREMENT NOT NULL,
  Title      VARCHAR(128) NOT NULL,
  Body       VARCHAR(255) NOT NULL,
  PRIMARY KEY (`PageID`)
);

INSERT INTO pages
  (Title, Body)
VALUES
  ('A', 'Hello world, my great friend!'),
  ('B', 'Hello world, my glorious friend!');