create keyspace tweetexample with replication = { 'class' : 'SimpleStrategy', 'replication_factor' : 1 };

create table tweetexample.tweet(timeline text, id UUID, text text, PRIMARY KEY(id));

create index on tweetexample.tweet(timeline);
