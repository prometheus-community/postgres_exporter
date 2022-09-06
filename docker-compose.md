In order to start PostgreSQL with SSL support, we need to change the file permissions
for the ssl cert and key.

Please run these commands if you want to run docker-compose locally.

sudo chown 999:999 testdata/ssl/server/*
sudo chmod 0600 testdata/ssl/server/*

Start the container:
```
docker-compose up
```

To be able to connect from pgsql you need to own the certs

sudo chown ${USER}:${USER} testdata/ssl/client*
sudo chmod 0600 testdata/ssl/client/*

Connect using psql

```
psql "host=127.0.0.1 port=5433 user=root password=root dbname=postgres sslmode=verify-ca sslcert=${PWD}/testdata/ssl/client/server.crt sslkey=${PWD}/testdata/ssl/client/server.key sslrootcert=${PWD}/testdata/ssl/client/CA.crt"
```

