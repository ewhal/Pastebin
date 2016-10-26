#Golang Pastebin
[![Build Status](https://travis-ci.org/ewhal/Pastebin.svg?branch=master)](https://travis-ci.org/ewhal/Pastebin) [![GoDoc](https://godoc.org/github.com/ewhal/Pastebin?status.svg)](https://godoc.org/github.com/ewhal/Pastebin) [![Go Report Card](https://goreportcard.com/badge/github.com/ewhal/Pastebin)](https://goreportcard.com/report/github.com/ewhal/Pastebin) [![MIT
licensed](https://img.shields.io/badge/license-MIT-blue.svg)](https://raw.githubusercontent.com/ewhal/Pastebin/master/LICENSE.md)

Modern self-hosted pastebin service with a restful API.

## Motivation
Many Pastebin services exist but all are more complicated than they need to be.
That is why I decided to write a pastebin service in golang.

![paste](http://i.imgur.com/7BeCKa3.png)

## Getting started
### Prerequisities
* pygmentize
* go
* mariadb

```
pip install pygmentize
sudo yum install -y go mariadb-server mariadb
```

### Installing
* Please note this assumes you have Mariadb and Go already setup.
* go get github.com/ewhal/Pastebin
* make
* mysql -u root -p
* CREATE USER 'paste'@'localhost' IDENTIFIED BY 'password';
* CREATE database paste;
* GRANT ALL PRIVILEGES ON paste . * TO 'paste'@'localhost';
* FLUSH PRIVILEGES;
* quit;
* mysql -u paste -p paste < database.sql
* cp config.example.json config.json
* nano config.json
* Configure port and database details

## License

This project is licensed under the MIT License - see the [LICENSE.md](LICENSE.md) file for details

