-- +goose Up

CREATE TABLE whitelist
(
  id INT PRIMARY KEY AUTO_INCREMENT,
  IP varchar(15) NOT NULL,
);

CREATE TABLE blacklist
(
  id INT PRIMARY KEY AUTO_INCREMENT,
  IP varchar(15) NOT NULL,
);


-- +goose Down
DROP TABLE whitelists;
DROP TABLE blacklists;