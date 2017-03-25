.mode csv
.separator , "\n"
.import all_funds.csv temp_funds
insert into funds(symbol, name, type) select * from temp_funds;
drop table temp_funds;
update funds set done = 0;
