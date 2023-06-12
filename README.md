# Neo4j simple Go bolt client

Happy path Go bolt client

### Local testing
```shell
docker run --publish=7474:7474 --publish=7687:7687 -e NEO4J_AUTH=neo4j/password -e NEO4J_ACCEPT_LICENSE_AGREEMENT=yes neo4j:4.3

go test
```
