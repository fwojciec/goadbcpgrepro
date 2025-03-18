Reproduction of a PostgreSQL ADBC driver SEGFAULT when used with Go FlightServer.

## Create Test Dataset

```sql
CREATE TABLE synthetic_filter_option (
  group_id    INTEGER,
  org_id      INTEGER,
  user_id     INTEGER,
  name        TEXT,
  code_type   TEXT,
  code        TEXT,
  description TEXT,
  code_id     INTEGER,
  archived    BOOLEAN
);
```

```sql
CREATE UNIQUE INDEX synthetic_fogmat ON synthetic_filter_option (org_id, group_id, code_id);
```

```sql
INSERT INTO synthetic_filter_option
SELECT
    (g % 100) AS group_id,                    -- cycles 0 to 99
    (g % 10) AS org_id,                       -- cycles 0 to 9
    (g % 1000) AS user_id,                    -- cycles 0 to 999
    'Group ' || g AS name,
    CASE WHEN g % 2 = 0 THEN 'TypeA' ELSE 'TypeB' END AS code_type,
    'Code' || g AS code,
    'Description ' || g AS description,
    g AS code_id,                             -- unique id per row
    (g % 2 = 0) AS archived                   -- even numbers true, odd false
FROM generate_series(1, 365712) AS s(g);
```
