#!/usr/bin/bash
docker run --name exams -e POSTGRES_PASSWORD=asusmhdsh -p 5432:5432 -d postgres:11-alpine