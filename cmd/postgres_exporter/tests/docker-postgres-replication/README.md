# Replicated postgres cluster in docker.

Upstream is forked from https://github.com/DanielDent/docker-postgres-replication

My version lives at https://github.com/wrouesnel/docker-postgres-replication

This very simple docker-compose file lets us stand up a replicated postgres
cluster so we can test streaming.

# TODO:
Pull in p2 and template the Dockerfile so we can test multiple versions.
