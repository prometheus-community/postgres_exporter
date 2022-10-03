## integration tests for exporter update

### Fast start:

run

    make prepare-env-from-repo

then run any of the ```make test-*```

### A bit of details:

1.  unpack original exporter


    make prepare-base-exporter

2.a. download updated exporter from specific feature build

    make prepare-exporter-from-fb url="<feature build client binary url>"

2.b. or use current repo as updated exporter

    make prepare-exporter-from-repo

3. start test postgres_server


    make start-postgres-db

4. run basic performance comparison test


    make test-performance

5.  run metrics list compatibility test


    make test-metrics

