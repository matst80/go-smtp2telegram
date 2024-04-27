# go-smtp2telegram

## Description

The go-smtp2telegram application is a tool that allows you to forward SMTP emails to Telegram. It provides a convenient way to receive email notifications directly on your Telegram account. It includes a webserver to be able to view html mails that is to complex for telegram and some basic spam protection

Quick hack to learn go, use with caution 😊

## Installation / usage

### Bare metal
To install and build the go-smtp2telegram application, follow these steps:

1. Clone the repository:

	```bash
	git clone git@github.com:matst80/go-smtp2telegram.git
	```

2. Change to the project directory:

	```bash
	cd go-smtp2telegram
	go mod download
	```

3. Build the application:

	```bash
	go build
	```

4. Create your configuration file config.json, copy example from config.example.json

5. Run the application:

	```bash
	./crapmail
	```

### Docker
Build the application or use [matst80/go-mail2telegram](https://hub.docker.com/r/matst80/go-mail2telegram)

```bash
docker build -t mail2telegram .
docker run -v ./config.json:/config.json -p 25:25 -p 8080:8080 mail2telegram
```

## Configuration

The config.json file is used to store configuration settings for your project. It typically contains various parameters that define the behavior and settings of your application. Let's take a look at some common parameters that you might find in a config.json file:

Remember to keep sensitive information, such as passwords or API keys, secure and avoid committing them to version control systems.

* `listen` address for smtp server default `0.0.0.0:25` 
* `domain` domain/hostname that the smtp server responds with
* `token` telegram bot token [instructions](https://core.telegram.org/bots/tutorial)
* `hashSalt` A string salt for basic user integrity (by no means safe) defaults to `salty-change-me`
* `baseUrl` webserver public address. f.ex. http://mail.domain.com
* `users` list of email and chatId objects, others will be discarded
* `stopWords` A list of blocking words that will be used for basic spam protection. 
* `warningWords` A list of words that increases the spam ranking
* `blockedIps` A list of blocked ips
* `warningWordsUrl` Url to download spam wordlist separated by \n
* `blockedIpUrl` Url to fetch updated spam classed ip numbers separated by \n

## Telegram commands

## `/start` 
will log the chatId on the server

## `/add user@domain.top other@domain.xyz`
Will add the the email addresses to the user that sends the message, sends the users list to be used to update the config.json file

## `/ips`
Updates the blocked ips list if a blockedIpUrl is provided in the config

## `/words`
Updates the warning word list if a blockedIpUrl is provided in the config

## `/config`
Replies the current configuration json

## `/users`
Replies the current users json
