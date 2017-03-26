
in the /db directory:
migrate up
`sqlite3 funds.db < import.sql`

Resources
====

 * [csv parsing in Go](https://www.dotnetperls.com/csv-go)


index_funds2 ranking query:
````select count(*) from funds;
select funds.symbol, funds.name, funds.type, avg(diff)*100 as performance_percent
from annual_returns
inner join funds on annual_returns.fund_id = funds.id
group by annual_returns.fund_id order by performance_percent desc
````
