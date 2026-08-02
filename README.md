![gophish logo](https://raw.github.com/gophish/gophish/main/static/images/gophish_purple.png)

Gophish
=======

![Build Status](https://github.com/gophish/gophish/workflows/CI/badge.svg) [![GoDoc](https://godoc.org/github.com/gophish/gophish?status.svg)](https://godoc.org/github.com/gophish/gophish)

Gophish: Open-Source Phishing Toolkit

[Gophish](https://getgophish.com) is an open-source phishing toolkit designed for businesses and penetration testers. It provides the ability to quickly and easily setup and execute phishing engagements and security awareness training.

### Install

Installation of Gophish is dead-simple - just download and extract the zip containing the [release for your system](https://github.com/gophish/gophish/releases/), and run the binary. Gophish has binary releases for Windows, Mac, and Linux platforms.

### Building From Source
**If you are building from source, please note that Gophish requires Go v1.26.5 or above!**

To build Gophish from source, simply run ```git clone https://github.com/gophish/gophish.git``` and ```cd``` into the project source directory. Then, run ```go build```. After this, you should have a binary called ```gophish``` in the current directory.

### Docker
You can also use Gophish via the official Docker container [here](https://hub.docker.com/r/gophish/gophish/).
The container starts without `config.json`; pass any overrides with `docker run -e NAME=value`.

### Environment Configuration

Environment variables override values from `config.json`. To run without a configuration file, pass an empty config path:

```bash
ADMIN_LISTEN_URL=127.0.0.1:3333 \
ADMIN_USE_TLS=false \
PHISH_LISTEN_URL=0.0.0.0:8080 \
PHISH_USE_TLS=false \
DB_NAME=sqlite3 \
DB_PATH=gophish.db \
MIGRATIONS_PREFIX=db/db_ \
./gophish --config=""
```

Comma-separated values are accepted for list settings.

| Environment variable | Configuration field |
| --- | --- |
| `ADMIN_LISTEN_URL` | `admin_server.listen_url` |
| `ADMIN_USE_TLS` | `admin_server.use_tls` |
| `ADMIN_CERT_PATH` | `admin_server.cert_path` |
| `ADMIN_KEY_PATH` | `admin_server.key_path` |
| `ADMIN_CSRF_KEY` | `admin_server.csrf_key` |
| `ADMIN_ALLOWED_INTERNAL_HOSTS` | `admin_server.allowed_internal_hosts` |
| `ADMIN_TRUSTED_ORIGINS` | `admin_server.trusted_origins` |
| `PHISH_LISTEN_URL` | `phish_server.listen_url` |
| `PHISH_USE_TLS` | `phish_server.use_tls` |
| `PHISH_CERT_PATH` | `phish_server.cert_path` |
| `PHISH_KEY_PATH` | `phish_server.key_path` |
| `DB_NAME` | `db_name` |
| `DB_PATH` | `db_path` (`DB_FILE_PATH` remains supported) |
| `DB_SSLCA_PATH` | `db_sslca_path` |
| `MIGRATIONS_PREFIX` | `migrations_prefix` |
| `CONTACT_ADDRESS` | `contact_address` |
| `LOGGING_FILENAME` | `logging.filename` |
| `LOGGING_LEVEL` | `logging.level` |
| `GOPHISH_INITIAL_ADMIN_PASSWORD` | Initial administrator password |
| `GOPHISH_INITIAL_ADMIN_API_TOKEN` | Initial administrator API token |

### Setup
After running the Gophish binary, open an Internet browser to https://localhost:3333 and login with the default username and password listed in the log output.
e.g.
```
time="2020-07-29T01:24:08Z" level=info msg="Please login with the username admin and the password 4304d5255378177d"
```

Releases of Gophish prior to v0.10.1 have a default username of `admin` and password of `gophish`.

### Documentation

Documentation can be found on our [site](http://getgophish.com/documentation). Find something missing? Let us know by filing an issue!

### Issues

Find a bug? Want more features? Find something missing in the documentation? Let us know! Please don't hesitate to [file an issue](https://github.com/gophish/gophish/issues/new) and we'll get right on it.

### License
```
Gophish - Open-Source Phishing Framework

The MIT License (MIT)

Copyright (c) 2013 - 2020 Jordan Wright

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software ("Gophish Community Edition") and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in
all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
THE SOFTWARE.
```
