WITH filtered_prices AS (
  SELECT
    fp.node_id,
    fp.fuel_type,
    fp.price_last_updated,
    fp.price,
    LAG(fp.price) OVER (
      PARTITION BY fp.node_id, fp.fuel_type
      ORDER BY fp.price_last_updated
    ) AS prev_price
  FROM petrol_filling_stations pfs
  INNER JOIN fuel_prices fp ON pfs.node_id = fp.node_id
  WHERE pfs.latitude BETWEEN ? AND ?
    AND pfs.longitude BETWEEN ? AND ?
),
price_changes AS (
  SELECT
    node_id,
    fuel_type,
    price_last_updated,
    price
  FROM filtered_prices
  WHERE price IS DISTINCT FROM prev_price
),
ranked_prices AS (
  SELECT
    node_id,
    fuel_type,
    price_last_updated,
    price,
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
  price
FROM ranked_prices
WHERE price_recency_rank <= ?;