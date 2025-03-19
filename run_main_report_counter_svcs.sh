#!/bin/bash

go run cmd/report-service/main.go

go run cmd/counter-service/main.go

go run cmd/main-service/main.go