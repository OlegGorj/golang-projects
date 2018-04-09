#!/bin/bash

cqlsh --color -u $1 -p $2 -f $PWD/cassandra-setup.cql $3
