
in the /db directory:
migrate up
`sqlite3 fund_crawler.db < import.sql`
`insert into funds(symbol, name, type) select * from temp_funds;`
