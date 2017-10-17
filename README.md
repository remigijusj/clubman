# Clubman

Web app for club scheduling.

Implemented in Go, using [gin-gonic](https://gin-gonic.github.io/gin/) web framework and SQLite DB.

## Features

* TOML configuration file
* Predefined languages, translations saved in DB
* Responsive [Foundation](http://foundation.zurb.com/)-based UI
* Simple email-based login
* Multiple teams containing members and instructor
* Events calendar, weekly and monthly
* Participant limits, waiting list for each event
* Self-signup, cancel or management by admin
* User notifications by email (through Mailgun) and SMS
* and more...

## Websites

* http://nk-fitness.dk

## Build & Run

`go build -o clubman ./src`

`./clubman 2>&1 | tee -a ./clubman.log`

## Compatibility

* Go v1.9.1
* Deps last updated 2017-10-17
