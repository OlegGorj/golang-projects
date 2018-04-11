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

### Example of config file:
```json
{
  "port": "8080",
  "serverslist": "127.0.0.1",
  "keyspace": "testapp",
  "userstable": "users",
  "sessionstable": "session",
  "username": "cassandra",
  "password": "cassandra"
}

```

External dependency: http://github.com/gocql/gocq

Logs will write to stdout.

## Summary application have two methods: /user/ and /session.

### Method `/user/`

`/user/` method realizing create user functionality. Accepting POST request with JSON structure like: `{ "username": "login", "password": "secret_password" }` and initializing adding such user to database. UPDATE/PUT and other request types will be ignored with "400 - Bad request" response.

`POST`
- 201 - Created -- when requested user created

- 400 - Bad request -- if had and error in JSON parsing, username is empty, password is empty or request in not POST.

- 409 - Conflict -- if user with such username was already created.

- 500 - Internal server error -- if errors happened on connection to database or proceeding CQL requests.

`GET`
Retrieves all users recorded in Cassandra DB with `active` flag NOT `false`. Responding with following codes:

200 - OK -- if request header is valid

401 - Unauthorized -- if user in header invalid or absent.

500 - Internal server error -- if errors happened on connection to database or proceeding CQL requests.

`DELETE`
Logically deleting user by setting `active` to `false`. Responding with following codes:

200 - OK -- if user deleted successfully. If user was absent, also response code is 200.

401 - Unauthorized -- if provided user is not valid.

500 - Internal server error -- if errors happened on connection to database or proceeding CQL requests.


### Method `/session/`

Accepting `POST`, `GET` or `DELETE` requests.

`POST`
Creating new session. Accepting JSON structure like { "username": "login", "password": "secret_password" }. Generating following responses:

201 - Created -- if session created successfully. Also adding "Set-Cookie: session_id=;" header to response. Cookie expire in one year, same TTL set to record in sessions table in Cassandra.

400 - Bad request -- if had and error in JSON parsing, username is empty, password is empty or request in not POST, GET or DELETE.

401 - Unauthorized -- if username or password is invalid.

500 - Internal server error -- if errors happened on connection to database or proceeding CQL requests.

`GET`
Checking session authorizing. Analyzing cookie session_id and verifying validity. Responding with following codes:

200 - OK -- if session_id cookie from request header is valid

401 - Unauthorized -- if session_id cookie in header invalid or absent.

500 - Internal server error -- if errors happened on connection to database or proceeding CQL requests.

`DELETE`
Deleting session indicated in cookie session_id. Responding with following codes:

200 - OK -- if session deleted successfully. If session_id cookie was absent, also response code is 200.

401 - Unauthorized -- if provided session_id is not valid.

500 - Internal server error -- if errors happened on connection to database or proceeding CQL requests.


## Database structure

Users table:
```sql
CREATE TABLE IF NOT EXISTS testapp.users (id UUID, username varchar, password varchar, active boolean, ts timestamp, PRIMARY KEY (id) )
```

Session table:
```sql
CREATE TABLE IF NOT EXISTS testapp.sessions (sessionID varchar PRIMARY KEY, username varchar)
```

Indexes to support lookups and queries

```sql
CREATE INDEX IF NOT EXISTS ON testapp.users (username)
CREATE INDEX IF NOT EXISTS ON testapp.users (active)
CREATE INDEX IF NOT EXISTS ON testapp.users (ts)
```

---
