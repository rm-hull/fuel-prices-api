-- Optimize fuel price stats by adding a covering index and rewriting the view
-- to join on station data AFTER filtering for latest prices.
-- Correctness is ensured by using a window function (ROW_NUMBER()) 
-- to accurately fetch the latest price per station/fuel pair.

CREATE INDEX IF NOT EXISTS idx_fuel_prices_latest 
ON fuel_prices(node_id, fuel_type, price_last_updated DESC, price);

DROP VIEW IF EXISTS fuel_price_snapshot_stats;

CREATE VIEW fuel_price_snapshot_stats AS
WITH latest_prices_ranked AS (
    -- Use a window function to find the latest price for each station/fuel combination
    -- before joining with other tables. This is much more efficient.
    SELECT
        node_id,
        fuel_type,
        price,
        ROW_NUMBER() OVER (PARTITION BY node_id, fuel_type ORDER BY price_last_updated DESC) as rn
    FROM fuel_prices
),
latest_snapshot AS (
    -- Now join the station data, but only for the latest prices (rn = 1).
    SELECT
        lpr.node_id,
        lpr.fuel_type,
        lpr.price,
        pfs.postcode
    FROM latest_prices_ranked lpr
    JOIN petrol_filling_stations pfs ON lpr.node_id = pfs.node_id
    WHERE lpr.rn = 1
),
with_postcode_area AS (
    -- Extract the postcode area (the leading alphabetic characters)
    SELECT
        fuel_type,
        price,
        UPPER(SUBSTR(TRIM(postcode), 1, LENGTH(TRIM(postcode)) - LENGTH(LTRIM(TRIM(postcode), 'ABCDEFGHIJKLMNOPQRSTUVWXYZ')))) as postcode_area
    FROM latest_snapshot
),
national_stats AS (
    SELECT
        'National' as scope,
        NULL as postcode_area,
        fuel_type,
        MIN(price) as min_price,
        ROUND(AVG(price),1) as avg_price,
        MAX(price) as max_price,
        SQRT(MAX(0, AVG(price * price) - AVG(price) * AVG(price))) as stddev_price,
        COUNT(*) as sample_size
    FROM with_postcode_area
    GROUP BY fuel_type
),
postcode_area_stats AS (
    SELECT
        'Postcode Area' as scope,
        postcode_area,
        fuel_type,
        MIN(price) as min_price,
        ROUND(AVG(price),1) as avg_price,
        MAX(price) as max_price,
        SQRT(MAX(0, AVG(price * price) - AVG(price) * AVG(price))) as stddev_price,
        COUNT(*) as sample_size
    FROM with_postcode_area
    WHERE postcode_area IS NOT NULL AND postcode_area <> ''
    GROUP BY postcode_area, fuel_type
)
SELECT * FROM national_stats
UNION ALL
SELECT * FROM postcode_area_stats;
