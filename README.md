# go-smtp2telegram

## Description

[Provide a brief description of your project here]

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

The `config.json` file is used to store configuration settings for your project. It is typically located in the root directory of your project. An example configuration file can be found in `config.example.json`.
