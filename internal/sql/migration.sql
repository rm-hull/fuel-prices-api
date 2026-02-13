-- SQLite migration for PetrolFillingStation and FuelPrice models
-- Creates tables to store stations and historical fuel price records.

PRAGMA foreign_keys = ON;

BEGIN TRANSACTION;

-- Stations table (maps to PetrolFillingStation)
CREATE TABLE IF NOT EXISTS petrol_filling_stations (
    node_id TEXT PRIMARY KEY,
    mft_organisation_name TEXT,
    public_phone_number TEXT,
    trading_name TEXT,
    is_same_trading_and_brand_name BOOLEAN NOT NULL DEFAULT 0,
    brand_name TEXT,
    temporary_closure BOOLEAN NOT NULL DEFAULT 0,
    permanent_closure BOOLEAN NOT NULL DEFAULT 0,
    permanent_closure_date TIMESTAMP,
    is_motorway_service_station BOOLEAN NOT NULL DEFAULT 0,
    is_supermarket_service_station BOOLEAN NOT NULL DEFAULT 0,
    address_line_1 TEXT,
    address_line_2 TEXT,
    city TEXT,
    country TEXT,
    county TEXT,
    postcode TEXT,
    latitude REAL,
    longitude REAL,
    opening_times_json TEXT,
    amenities_json TEXT,
    fuel_types_json TEXT,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX IF NOT EXISTS idx_petrol_lat_lng ON petrol_filling_stations(latitude, longitude);

-- Historical fuel prices (maps to FuelPrice; supports many records per station/fuel)
CREATE TABLE IF NOT EXISTS fuel_prices (
    node_id TEXT NOT NULL,
    fuel_type TEXT NOT NULL,
    price_last_updated TIMESTAMP NOT NULL,
    price REAL NOT NULL,
    recorded_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (node_id, fuel_type, recorded_at),
    FOREIGN KEY (node_id) REFERENCES petrol_filling_stations(node_id) ON DELETE CASCADE
);


COMMIT;
