SELECT
  node_id,
  fuel_type,
  price_last_updated,
  price
FROM (
  SELECT
    node_id,
    fuel_type,
    price_last_updated,
    price,
    ROW_NUMBER() OVER (PARTITION BY node_id, fuel_type ORDER BY price_last_updated DESC) AS price_recency_rank
  FROM (
    SELECT
      fp.node_id,
      fp.fuel_type,
      fp.price_last_updated,
      fp.price,
      ROW_NUMBER() OVER (PARTITION BY fp.node_id, fp.fuel_type, fp.price ORDER BY fp.price_last_updated DESC) AS same_price_rank
    FROM petrol_filling_stations pfs
    INNER JOIN fuel_prices fp ON pfs.node_id = fp.node_id
    WHERE pfs.latitude BETWEEN ? AND ?
      AND pfs.longitude BETWEEN ? AND ?
  ) t1
  WHERE same_price_rank = 1
) t2
WHERE price_recency_rank <= ?;