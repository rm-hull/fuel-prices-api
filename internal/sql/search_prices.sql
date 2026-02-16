SELECT
  node_id,
  fuel_type,
  price_last_updated,
  price
FROM (
  SELECT
    fp.node_id,
    fp.fuel_type,
    fp.price_last_updated,
    fp.price,
    ROW_NUMBER() OVER (PARTITION BY fp.node_id, fp.fuel_type ORDER BY fp.price_last_updated DESC) AS rn
  FROM petrol_filling_stations pfs
  INNER JOIN fuel_prices fp ON pfs.node_id = fp.node_id
  WHERE pfs.latitude BETWEEN ? AND ?
    AND pfs.longitude BETWEEN ? AND ?
) t
WHERE rn <= ?;