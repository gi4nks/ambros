# ambros
The command butler!! 

[![Travis CI Build Status](https://travis-ci.org/gi4nks/ambros.svg?branch=master)](https://travis-ci.org/gi4nks/ambros)

ambros creates a local history of executed commands, keeipng track also of the output. At the monet it does not work with interactive commands.

## how to install

> go get github.com/gi4nks/ambros

## how to build

> go build -o bin/ambros cmd/main.go

## how to use

### run a command

> ambros ru -- ls -la

### getting help

> ambros help