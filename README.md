
in the /db directory:
migrate up
`sqlite3 funds.db < import.sql`

Resources
====

 * [csv parsing in go](https://www.dotnetperls.com/csv-go)
 * [common go mistakes](http://devs.cloudimmunity.com/gotchas-and-common-mistakes-in-go-golang/)
 * [channels are not enough](https://gist.github.com/kachayev/21e7fe149bc5ae0bd878)
 * [when to use pointers](http://stackoverflow.com/questions/23542989/pointers-vs-values-in-parameters-and-return-values)
 * [go interfaces](https://research.swtch.com/interfaces)
 * [go memory pools](https://blog.cloudflare.com/recycling-memory-buffers-in-go/)
 * [go style guide - receiver names (READ ALL OF THIS)](https://github.com/golang/go/wiki/CodeReviewComments#pass-values)


index_funds2 ranking query:
````select count(*) from funds;
select funds.symbol, funds.name, funds.type, avg(diff)*100 as performance_percent
from annual_returns
inner join funds on annual_returns.fund_id = funds.id
group by annual_returns.fund_id order by performance_percent desc
````
