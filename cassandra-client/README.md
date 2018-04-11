[![GitHub release](https://img.shields.io/github/release/OlegGorj/go-templates-collection.svg)](https://github.com/OlegGorj/go-templates-collection/releases)
[![GitHub issues](https://img.shields.io/github/issues/OlegGorj/go-templates-collection.svg)](https://github.com/OlegGorj/go-templates-collection/issues)
[![Codacy Badge](https://api.codacy.com/project/badge/Grade/1818748c6ba745ce97bb43ab6dbbfd2c)](https://www.codacy.com/app/oleggorj/go-templates-collection?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=OlegGorj/go-templates-collection&amp;utm_campaign=Badge_Grade)

# cassandra-go-client

Very light and simple Cassandra client written in GO for testing purposes.

## Overview

What it does:

1. Makes connection to Cassandra cluster. Keyspace `example` and table `tweet`.
2. Inserts test record to table `tweet`
3. Selects first record `id` and `text` from table `tweet`
4. Retrieves all records produced by `me` from the table `tweet` and constructs collection of slices - one record per slice
5. Prints all slices to stdout


## Cassandra setup steps

There are a few configuration tweaks required for the code to work properly

Authentication Setup:

1. in `cassandra.yml`, set `authenticator` as `PasswordAuthenticator`. This is done to support password-based authentication.
2. again, in `cassandra.yml`, set `authorizer` as `CassandraAuthorizer`. In order for authorization to work properly, we require use of `CassandraAuthorizer`. If, for your purposes you don't need authorization at all, you can switch to use `AllowAllAuthorizer`, which disables authorization.

For more details related to cassandra.yml and configurations, here is [the link](https://docs.datastax.com/en/cassandra/3.0/cassandra/configuration/configCassandra_yaml.html)
For more detailed information, refer to http://docs.datastax.com/en/cassandra/3.0/cassandra/security/security_config_native_authenticate_t.html.

## Cassandra keyspace and table setup steps


Then, execute the following commands to create keyspace, table and index.
You would need to provide user name and password of Admin type of user, as well as IP of your cluster.
Let's assume for now, user name is `cassandra`, passwords is `cassandra` and IP is local hosts `127.0.0.1`

```
git clone https://github.com/OlegGorj/go-templates-collection.git
cd ./go-templates-collection/cassandra-client
./cassandra-setup.sh cassandra cassandra 127.0.0.1
```
last command should complete with no messages.


Using `cqlsh` login to your C* cluster

```
cqlsh -u <username> -p <password>
```

To test your setup, insert sample record:

```
INSERT INTO example.tweet (timeline, id, text) VALUES ('me', UUID(), 'hello world');
SELECT * FROM example.tweet;
```

## Cassandra client use

```
go run cassandra-client.go -h 127.0.0.1 -u cassandra -p cassandra
```


---
