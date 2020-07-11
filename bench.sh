#!/usr/bin/env bash
go test -v -benchmem -bench=.* -benchtime 2s
