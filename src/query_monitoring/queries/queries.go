// Package queries contains the collection methods to parse and build the collection schema
package queries

const (
	SlowQueries = `SELECT
        pss.queryid AS query_id,
        pss.query AS query_text,
        pd.datname AS database_name,
        current_schema() AS schema_name,
        pss.calls AS execution_count,
        ROUND((pss.total_exec_time / pss.calls)::numeric, 3) AS avg_elapsed_time_ms,
        ROUND((pss.total_exec_time / pss.calls)::numeric, 3) AS avg_cpu_time_ms,
        pss.shared_blks_read / pss.calls AS avg_disk_reads,
        pss.shared_blks_written / pss.calls AS avg_disk_writes,
        CASE
            WHEN pss.query ILIKE 'SELECT%' THEN 'SELECT'
            WHEN pss.query ILIKE 'INSERT%' THEN 'INSERT'
            WHEN pss.query ILIKE 'UPDATE%' THEN 'UPDATE'
            WHEN pss.query ILIKE 'DELETE%' THEN 'DELETE'
            ELSE 'OTHER'
        END AS statement_type,
        to_char(NOW() AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS collection_timestamp
    FROM
        pg_stat_statements pss
    JOIN
        pg_database pd ON pss.dbid = pd.oid
    ORDER BY
        avg_elapsed_time_ms DESC -- Order by the average elapsed time in descending order
    LIMIT
        5;`

	WaitEvents = `WITH wait_history AS (
        SELECT
            wh.pid,
            wh.event_type,
            wh.event,
            wh.ts,
            pg_database.datname AS database_name,
            LEAD(wh.ts) OVER (PARTITION BY wh.pid ORDER BY wh.ts) - wh.ts AS duration,
            sa.query AS query_text,
            sa.queryid AS query_id
        FROM
            pg_wait_sampling_history wh
        LEFT JOIN
            pg_stat_statements sa ON wh.queryid = sa.queryid
        LEFT JOIN
            pg_database ON pg_database.oid = sa.dbid
    )
    SELECT
        event_type || ':' || event AS wait_event_name,
        CASE
            WHEN event_type IN ('LWLock', 'Lock') THEN 'Locks'
            WHEN event_type = 'IO' THEN 'Disk IO'
            WHEN event_type = 'CPU' THEN 'CPU'
            ELSE 'Other'
        END AS wait_category,
        EXTRACT(EPOCH FROM SUM(duration)) * 1000 AS total_wait_time_ms,  -- Convert duration to milliseconds
        COUNT(*) AS waiting_tasks_count,
        to_char(NOW() AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"') AS collection_timestamp,
        query_id,
        query_text,
        database_name
    FROM wait_history
    WHERE duration IS NOT NULL AND query_id IS NOT NULL AND event_type IS NOT NULL
    GROUP BY event_type, event, query_id, query_text, database_name
    ORDER BY total_wait_time_ms DESC
    LIMIT 10;`
	BlockingQueries = `SELECT
    blocked_locks.pid AS blocked_pid,
    blocked_activity.usename AS blocked_user,
    blocking_locks.pid AS blocking_pid,
    blocking_activity.usename AS blocking_user,
    blocked_activity.query AS blocked_query,
    blocking_activity.query AS blocking_query
FROM pg_catalog.pg_locks blocked_locks
JOIN pg_catalog.pg_stat_activity blocked_activity ON blocked_activity.pid = blocked_locks.pid
JOIN pg_catalog.pg_locks blocking_locks ON blocking_locks.locktype = blocked_locks.locktype
    AND blocking_locks.DATABASE IS NOT DISTINCT FROM blocked_locks.DATABASE
    AND blocking_locks.relation IS NOT DISTINCT FROM blocked_locks.relation
    AND blocking_locks.page IS NOT DISTINCT FROM blocked_locks.page
    AND blocking_locks.tuple IS NOT DISTINCT FROM blocked_locks.tuple
    AND blocking_locks.virtualxid IS NOT DISTINCT FROM blocked_locks.virtualxid
    AND blocking_locks.transactionid IS NOT DISTINCT FROM blocked_locks.transactionid
    AND blocking_locks.classid IS NOT DISTINCT FROM blocked_locks.classid
    AND blocking_locks.objid IS NOT DISTINCT FROM blocked_locks.objid
    AND blocking_locks.objsubid IS NOT DISTINCT FROM blocked_locks.objsubid
JOIN pg_catalog.pg_stat_activity blocking_activity ON blocking_activity.pid = blocking_locks.pid
WHERE NOT blocked_locks.GRANTED;
`
)
