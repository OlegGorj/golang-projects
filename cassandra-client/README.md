# cassandra-client

Very light and simple Cassandra client written in GO for testing purposes.

## Overview

What it does:

1. Makes connection to Cassandra cluster. Keyspace `example` and table `tweet`.
2. Inserts test record to table `tweet`
3. Selects first record `id` and `text` from table `tweet`
4. Retrieves all records produced by `me` from the table `tweet` and constructs collection of slices - one record per slice
5. Prints all slices to stdout


## Cassandra setup steps

There are a few configuration tweaks required for the code to work properly:

1. in `cassandra.yml`, set `authenticator` as `PasswordAuthenticator`. This is done to support password-based authentication.
2. again, in `cassandra.yml`, set `authorizer` as `CassandraAuthorizer`. In order for authorization to work properly, we require use of `CassandraAuthorizer`. If, for your purposes you don't need authorization at all, you can switch to use `AllowAllAuthorizer`, which disables authorization.

For more details related to cassandra.yml and configurations, here is [the link](https://docs.datastax.com/en/cassandra/3.0/cassandra/configuration/configCassandra_yaml.html)


## Cassandra keyspace and table setup steps

Using `cqlsh` login to your C* cluster

```
cqlsh -u cassandra -p cassandra
```

Then, execute the following commands to create keyspace, table and index.

```
create keyspace example with replication = { 'class' : 'SimpleStrategy', 'replication_factor' : 1 };
create table example.tweet(timeline text, id UUID, text text, PRIMARY KEY(id));
create index on example.tweet(timeline);
```

To test your setup, insert sample record:

```
INSERT INTO example.tweet (timeline, id, text) VALUES ('me', UUID(), 'hello world');
```



---
