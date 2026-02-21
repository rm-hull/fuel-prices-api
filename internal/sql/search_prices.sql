WITH filtered_prices AS (
  SELECT
    fp.node_id,
    fp.fuel_type,
    fp.price_last_updated,
    fp.price,
    fp.price_change_effective_timestamp,
    LAG(fp.price) OVER (
      PARTITION BY fp.node_id, fp.fuel_type
      ORDER BY fp.price_last_updated
    ) AS prev_price
  FROM petrol_filling_stations pfs
  INNER JOIN fuel_prices fp ON pfs.node_id = fp.node_id
  WHERE pfs.latitude BETWEEN ? AND ?
    AND pfs.longitude BETWEEN ? AND ?
),
grouped AS (
  SELECT
    node_id,
    fuel_type,
    price,
    price_change_effective_timestamp,
    price_last_updated,
      SUM(CASE WHEN price IS DISTINCT FROM prev_price THEN 1 ELSE 0 END)
        OVER (PARTITION BY node_id, fuel_type ORDER BY price_last_updated ROWS UNBOUNDED PRECEDING) AS grp
  FROM filtered_prices
),
price_changes AS (
  -- For each run of identical prices pick the most recent row for that run
  SELECT
    node_id,
    fuel_type,
    price_last_updated,
    price,
    price_change_effective_timestamp
  FROM (
    SELECT *, ROW_NUMBER() OVER (PARTITION BY node_id, fuel_type, grp ORDER BY price_last_updated DESC) AS rn
    FROM grouped
  ) t
  WHERE rn = 1
),
ranked_prices AS (
  SELECT
    node_id,
    fuel_type,
    price_last_updated,
    price,
    price_change_effective_timestamp,
    ROW_NUMBER() OVER (
      PARTITION BY node_id, fuel_type
      ORDER BY price_last_updated DESC
    ) AS price_recency_rank
  FROM price_changes
)
SELECT
  node_id,
  fuel_type,
  price_last_updated,
  price,
  price_change_effective_timestamp
FROM ranked_prices
WHERE price_recency_rank <= ?;