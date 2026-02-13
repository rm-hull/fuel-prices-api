SELECT
    fp.node_id,
    fp.fuel_type,
    fp.price_last_updated,
    fp.price
FROM petrol_filling_stations pfs
INNER JOIN fuel_prices fp ON pfs.node_id = fp.node_id
WHERE pfs.latitude BETWEEN ? AND ?
  AND pfs.longitude BETWEEN ? AND ?;