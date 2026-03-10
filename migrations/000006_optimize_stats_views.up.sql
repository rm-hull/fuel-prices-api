-- Optimize both stats views using MATERIALIZED CTEs to avoid redundant processing
-- and ensuring joins happen AFTER filtering for the latest prices.

-- 1. Optimize snapshot stats
DROP VIEW IF EXISTS fuel_price_snapshot_stats;
CREATE VIEW fuel_price_snapshot_stats AS
WITH latest_prices_ranked AS MATERIALIZED (
    SELECT
        node_id,
        fuel_type,
        price,
        ROW_NUMBER() OVER (PARTITION BY node_id, fuel_type ORDER BY price_last_updated DESC) as rn
    FROM fuel_prices
),
latest_snapshot AS MATERIALIZED (
    SELECT
        lpr.node_id,
        lpr.fuel_type,
        lpr.price,
        pfs.postcode
    FROM latest_prices_ranked lpr
    JOIN petrol_filling_stations pfs ON lpr.node_id = pfs.node_id
    WHERE lpr.rn = 1
),
with_postcode_area AS MATERIALIZED (
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

-- 2. Optimize distribution stats
DROP VIEW IF EXISTS fuel_price_distribution_stats;
CREATE VIEW fuel_price_distribution_stats AS
WITH latest_prices_ranked AS MATERIALIZED (
    SELECT
        node_id,
        fuel_type,
        price,
        ROW_NUMBER() OVER (PARTITION BY node_id, fuel_type ORDER BY price_last_updated DESC) as rn
    FROM fuel_prices
),
latest_snapshot AS MATERIALIZED (
    SELECT
        lpr.node_id,
        lpr.fuel_type,
        lpr.price,
        pfs.postcode
    FROM latest_prices_ranked lpr
    JOIN petrol_filling_stations pfs ON lpr.node_id = pfs.node_id
    WHERE lpr.rn = 1
),
with_postcode_area AS MATERIALIZED (
    SELECT
        fuel_type,
        price,
        UPPER(SUBSTR(TRIM(postcode), 1, LENGTH(TRIM(postcode)) - LENGTH(LTRIM(TRIM(postcode), 'ABCDEFGHIJKLMNOPQRSTUVWXYZ')))) as postcode_area
    FROM latest_snapshot
),
national_distribution AS (
    SELECT
        'National' as scope,
        NULL as postcode_area,
        fuel_type,
        (CAST(price / 2 AS INT) * 2) as price_bucket,
        COUNT(*) as sample_size
    FROM with_postcode_area
    GROUP BY fuel_type, price_bucket
),
postcode_area_distribution AS (
    SELECT
        'Postcode Area' as scope,
        postcode_area,
        fuel_type,
        (CAST(price / 2 AS INT) * 2) as price_bucket,
        COUNT(*) as sample_size
    FROM with_postcode_area
    WHERE postcode_area IS NOT NULL AND postcode_area <> ''
    GROUP BY postcode_area, fuel_type, price_bucket
)
SELECT * FROM national_distribution
UNION ALL
SELECT * FROM postcode_area_distribution;
