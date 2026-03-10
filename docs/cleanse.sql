-- Some historical data looks like it has been keyed in incorrectly at some point.
-- This ad-hoc script will either adjust or delete the most egregious errors.
-- It may need to be run occasionally to fix bad data.
delete from fuel_prices where price = 999.9;
delete from fuel_prices where price = 999.99;
delete from fuel_prices where price = 0.999;
delete from fuel_prices where price = 99.9;
update fuel_prices set price = price * 100 where price >=1 and price <2;
update fuel_prices set price = price * 10 where price >=12 and price <20;
update fuel_prices set price = price + 100 where price >=20 and price <100;
delete from fuel_prices where price < 100;

