-- Optimize fuel price stats by adding a covering index and rewriting the view
-- to join on station data AFTER filtering for latest prices.

CREATE INDEX IF NOT EXISTS idx_fuel_prices_latest 
ON fuel_prices(node_id, fuel_type, price_last_updated DESC, price);

DROP VIEW IF EXISTS fuel_price_snapshot_stats;

CREATE VIEW fuel_price_snapshot_stats AS
WITH latest_prices_only AS (
    -- Efficiently find the latest price per station/fuel using the index
    SELECT node_id, fuel_type, MAX(price_last_updated) as latest_update, price
    FROM fuel_prices
    GROUP BY node_id, fuel_type
),
latest_snapshot AS (
    -- ONLY join station data for the latest rows
    SELECT 
        lpo.node_id, 
        lpo.fuel_type, 
        lpo.price, 
        pfs.postcode
    FROM latest_prices_only lpo
    JOIN petrol_filling_stations pfs ON lpo.node_id = pfs.node_id
),
with_postcode_area AS (
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
