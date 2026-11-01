---
title: Running Against AWS RDS
nav_title: AWS RDS
sort_rank: 8
---
# Running Against AWS RDS

RDS is a managed PostgreSQL service, so you don't have superuser access — a few things need to be set up differently than on a self-managed server.

## Excluded database and role

`rdsadmin` is an RDS-internal database and role you can't (and shouldn't) query. Point `DATA_SOURCE_NAME` at your actual database (e.g. `postgres`), same as any other target — see [Connecting to PostgreSQL](connecting.md). If you use the deprecated `--auto-discover-databases` flag to scrape every database on the instance, exclude it with `PG_EXPORTER_EXCLUDE_DATABASES=rdsadmin`.

## Permissions

Follow the standard non-superuser [database permissions](database-permissions.md) setup, granting `pg_monitor` to your exporter role. Since you don't have a true superuser on RDS, run the grant as your RDS master user, and grant the exporter role to the master user first if needed (see the note in [database-permissions.md](database-permissions.md#postgresql-10-use-the-built-in-pg_monitor-role)).

## `pg_stat_statements`

To use the `stat_statements` collector, `pg_stat_statements` must be preloaded via the RDS parameter group, not just created as an extension:

1. In the RDS parameter group attached to your instance, ensure that the pg_stat_statements extension is preloaded:
   ```
   shared_preload_libraries = "pg_stat_statements"
   ```
2. Reboot the RDS instance for the change to take effect.
3. Then `CREATE EXTENSION pg_stat_statements;` as usual.

`pg_stat_statements` rows include queries run by RDS's own `rdsadmin` role. Exclude them with `--collector.stat_statements.exclude_users=rdsadmin`.

## IAM database authentication

With [IAM database authentication](https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/UsingWithRDS.IAMDBAuth.html), the exporter doesn't need a database password. For each new connection it signs a short-lived authentication token with its AWS credentials and uses the token as the password. RDS for PostgreSQL and Aurora PostgreSQL both support it.

### Setting it up

1. [Enable IAM database authentication](https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/UsingWithRDS.IAMDBAuth.Enabling.html) on the instance or cluster.
2. Create the exporter's database user, grant it `rds_iam`, and give it the [usual permissions](database-permissions.md):
   ```sql
   CREATE USER postgres_exporter;
   GRANT rds_iam TO postgres_exporter;
   GRANT pg_monitor TO postgres_exporter;
   ```
   A user with `rds_iam` can no longer log in with a password, so don't grant it to the master user.
3. Allow the exporter's IAM identity to connect as that user, with an [IAM policy](https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/UsingWithRDS.IAMDBAuth.IAMPolicy.html) for `rds-db:connect`:
   ```json
   {
     "Version": "2012-10-17",
     "Statement": [
       {
         "Effect": "Allow",
         "Action": "rds-db:connect",
         "Resource": "arn:aws:rds-db:us-east-1:123456789012:dbuser:cluster-ABCDEFGHIJKL01234/postgres_exporter"
       }
     ]
   }
   ```
   The resource uses the resource ID of the Aurora cluster (`cluster-…`) or RDS instance (`db-…`), not its name.
4. Give the exporter AWS credentials. It uses the AWS SDK's default chain: environment variables, the shared config and credentials files, web identity (IAM roles for service accounts on EKS), container credentials (ECS task roles, EKS Pod Identity), or the EC2 instance profile. The region comes from `AWS_REGION` or the shared config; on EC2 it falls back to the instance's region.

### Single target (`/metrics`)

Set `--datasource.auth=iam` and leave the password out:

```bash
docker run \
  -e DATA_SOURCE_USER=postgres_exporter \
  -e AWS_REGION=us-east-1 \
  -v /path/to/global-bundle.pem:/etc/rds/global-bundle.pem:ro \
  quay.io/prometheuscommunity/postgres-exporter \
  --datasource.auth=iam \
  --datasource.uri="mydb.cluster-abcdefghijkl.us-east-1.rds.amazonaws.com:5432/postgres?sslmode=verify-full&sslrootcert=/etc/rds/global-bundle.pem"
```

### Multi-target (`/probe`)

Use an `iam` auth module in the [config file](configuration.md#config-file). Both `region` and `role_arn` are optional: without them the module uses the exporter's own region and credentials, and with `role_arn` it assumes that role to sign the tokens.

```yaml
auth_modules:
  orders:
    type: iam
    iam:
      region: eu-west-1
      role_arn: arn:aws:iam::123456789012:role/orders-exporter
    options:
      user: postgres_exporter
      sslmode: verify-full
      sslrootcert: /etc/rds/global-bundle.pem
```

### Things to know

- The connection must use TLS: `sslmode` has to be `require` (the default), `verify-ca` or `verify-full`. The exporter refuses `disable`, `allow` and `prefer`, which could send the token in plaintext. Use `verify-full` with the [RDS certificate bundle](https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/UsingWithRDS.SSL.html) to also check the server's identity.
- Connect to the RDS endpoint itself. The token is signed for the host name you connect to, so a custom DNS name, such as a Route 53 CNAME, doesn't work.
- The data source must name a single host and must not contain a password. The exporter refuses to start otherwise.
- Without a region, every scrape fails with `no AWS region configured; set iam.region or AWS_REGION`.
- AWS recommends 300 to 1000 MiB of extra memory on the database for IAM authentication; see its [limitations](https://docs.aws.amazon.com/AmazonRDS/latest/AuroraUserGuide/UsingWithRDS.IAMDBAuth.html).
