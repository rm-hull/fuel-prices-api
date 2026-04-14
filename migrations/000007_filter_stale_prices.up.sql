-- Filter out stale prices (older than 14 days) from the stats views.

DROP VIEW IF EXISTS fuel_price_latest_with_area;
CREATE VIEW fuel_price_latest_with_area AS
WITH latest_prices_ranked AS (
    SELECT
        node_id,
        fuel_type,
        price,
        ROW_NUMBER() OVER (PARTITION BY node_id, fuel_type ORDER BY price_last_updated DESC) as rn
    FROM fuel_prices
    WHERE price_last_updated >= datetime('now', '-14 days')
),
latest_snapshot AS (
    SELECT
        lpr.node_id,
        lpr.fuel_type,
        lpr.price,
        pfs.postcode
    FROM latest_prices_ranked lpr
    JOIN petrol_filling_stations pfs ON lpr.node_id = pfs.node_id
    WHERE lpr.rn = 1
)
SELECT
    fuel_type,
    price,
    UPPER(SUBSTR(TRIM(postcode), 1, LENGTH(TRIM(postcode)) - LENGTH(LTRIM(TRIM(postcode), 'ABCDEFGHIJKLMNOPQRSTUVWXYZ')))) as postcode_area
FROM latest_snapshot;
