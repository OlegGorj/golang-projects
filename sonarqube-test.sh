#!/bin/bash

~/Downloads/sonar-scanner-3.1.0.1141-macosx/bin/sonar-scanner \
  -Dsonar.projectKey=cassandra-client \
  -Dsonar.organization=oleggorj-github \
  -Dsonar.sources=. \
  -Dsonar.host.url=https://sonarcloud.io \
  -Dsonar.login=34bd40c5683853e6ddd7dfabb636af6f064a7a4b
