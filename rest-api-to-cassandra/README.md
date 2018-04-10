[![GitHub release](https://img.shields.io/github/release/OlegGorj/go-templates-collection.svg)](https://github.com/OlegGorj/go-templates-collection/releases)
[![GitHub issues](https://img.shields.io/github/issues/OlegGorj/go-templates-collection.svg)](https://github.com/OlegGorj/go-templates-collection/issues)
![Quality Gates](https://sonarcloud.io/api/project_badges/measure?project=cassandra-client&metric=alert_status)
[![Codacy Badge](https://api.codacy.com/project/badge/Grade/1818748c6ba745ce97bb43ab6dbbfd2c)](https://www.codacy.com/app/oleggorj/go-templates-collection?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=OlegGorj/go-templates-collection&amp;utm_campaign=Badge_Grade)

# rest-api-to-cassandra

This project designed to test functionality of Golang REST API with Cassandra as backend.

Service have configuration file(config.json), which contains:

 - port for service listen
 - addresses of Cassandra servers
 - Keyspace name. If such keyspace is absent, application will create new one with necessary tables.

Example of config file:
```json
{
  "port": "8080",
  "serverslist": "127.0.0.1,127.0.0.1",
  "keyspace": "testapp",
  "username": "cassandra",
  "password": "cassandra"
}

```

External dependency: http://github.com/gocql/gocq

Logs will write to stdout.
