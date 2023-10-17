![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/CGI-FR/LINO/ci.yml?branch=main)
[![Go Report Card](https://goreportcard.com/badge/github.com/cgi-fr/pimo)](https://goreportcard.com/report/github.com/cgi-fr/lino)
![GitHub all releases](https://img.shields.io/github/downloads/CGI-FR/LINO/total)
![GitHub](https://img.shields.io/github/license/CGI-FR/LINO)
![GitHub Repo stars](https://img.shields.io/github/stars/CGI-FR/LINO)
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/CGI-FR/LINO)
![GitHub release (latest by date)](https://img.shields.io/github/v/release/CGI-FR/LINO)

# LINO : Large Input, Narrow Output

LINO is a simple ETL (Extract Transform Load) tools to manage tests datas.
The `lino` command line tool pull test data from a relational database to create a smallest production-like database.

![lino](doc/img/lino.png)

## Usage

`lino` command line work in relative project's directory, like `git` or `docker`

The following flags can be used:

* `--verbosity <level>` or `-v<level>` This flag increase verbosity on the stderr output, possible values: none (0), error (1), warn (2), info (3), debug (4), trace (5).
* `--debug` This flag complete the logs with debug information.
* `--log-json` Set this flag to produce JSON formatted logs goes deeper into logging and structured logging)
* `color` This flag allows to add color to console logs
* `--stats <filename | url>` This flag either outputs run statistics to the specified file or send them to specified url (has to start with `http` or `https`).
* `--statsTemplate <string>` This flag will have LINO use the value as a template to generate statistics. Please use go templating format to include statistics. To include them you have to specify them as `{{ .Stats }}`. (i.e. `{"software":"LINO","stats":{{ .Stats }}}`

## Create a new LINO project

```
$ mkdir myproject
$ cd myproject
```

## Add DataConnector

A DataConnector is a database connection shortcut.

```bash
$ lino dataconnector add source postgresql://postgres:sakila@localhost:5432/postgres?sslmode=disable
successfully added dataconnector {source postgresql://postgres:sakila@localhost:5432/postgres?sslmode=disable}
```

### Connection string

Lino use a connection string following an URL Schema.

```
<databaseVendor>://[<user>[:<password>]]@<host>/<database>
```

Currently supported vendors are :

* postgresql
* oracle
* oracle-raw (for full TNS support `oracle-raw://user:pwd@(DESCRIPTION=(ADDRESS=(PROTOCOL=TCP)(HOST=dbhost.example.com)(PORT=1521))(CONNECT_DATA=(SERVICE_NAME=orclpdb1)))`)
* mysql / mariadb
* db2 (alpha feature) : the DB2 driver is currently in development, contact us for a compilation of a LINO binary with DB2 support with your target os/arch
* http : use an HTTP endpoint to push and pull data (for databases with no native driver supported by golang)

### dataconnector.yml

The content of `dataconnector.yml` generated is

```yaml
version: v1
dataconnectors:
  - name: source
    url: postgresql://postgres:sakila@localhost:5432/postgres?sslmode=disable
```

## Create relationships

LINO create a consistent sample database. To perform extraction that respect foreign keys constraints LINO have to extract relationships between tables.

Use the `relation` sub-command or its short name `rel` to extract relations from foreign key constraints.


```
$ lino relation extract source
lino finds 40 relations from constraints
```

The content of `relations.yml` generated is

```yaml
version: v1
relations:
  - name: film_original_language_id_fkey
    parent:
        name: public.film
        keys:
          - original_language_id
    child:
        name: public.language
        keys:
          - language_id
  - name: film_language_id_fkey
.
.
.
```

At least user can edit the `relations.yml` manually to add relations that are not part of the database model.

## Extract Tables

The `table` action extract informations about tables.

```
$ lino table extract source
lino finds 15 table(s)
```

`lino` store the table description in `table.yml` file :

```yaml
version: v1
tables:
  - name: public.actor
    keys:
      - actor_id
  - name: public.address
    keys:
      - address_id
  - name: public.category
    keys:
```

## Ingress descriptor

Ingress descriptor object describe how `lino` has to go through the relations to extract data test.

### Create Ingress descriptor

To create ingress descriptor use the `id` sub-command with the start table of the extraction.

```bash
$ lino id create public.customer
successfully created ingress descriptor
```

`lino` store the new ingress descriptor in `ingress-descriptor.yml` file :

```yaml
version: v1
IngressDescriptor:
    startTable: public.customer
    relations:
      - name: film_original_language_id_fkey
        parent:
            name: public.film
            lookup: false
        child:
            name: public.language
            lookup: true
```

Customize the extraction plan of the ingress descriptor by editing the `ingress-descriptor.yml` file or by using the dedicated commands.

For example, this version of the `ingress-descriptor.yml` will filter out language objects created before 01/01/2023 :

```yaml
version: v1
IngressDescriptor:
    startTable: public.customer
    relations:
      - name: film_original_language_id_fkey
        parent:
            name: public.film
            lookup: false
        child:
            name: public.language
            lookup: true
            where: "creation_date >= '01/01/2023'"
```

The `where` property can be set on the `child` or the `parent` object. When used on the parent object, extracted data will contains a null parent object if the parent is filtered by the where clause. When used on chlid object, the resulting child list will be filtered (if all children are filtered, the list will be empty).

To modify the `ingress-descriptor.yml`, some commands can be used instead of editing directly the file :
- `lino id set-child-lookup <relation name> <true or false>` : modify the `lookup` property of the child object
- `lino id set-parent-lookup <relation name> <true or false>` : modify the `lookup` property of the parent object
- `lino id set-child-where <relation name> <where clause>` : modify the `where` property of the child object
- `lino id set-parent-where <relation name> <where clause>` : modify the `where` property of the parent object
- `lino id set-start-table <table name>` : modify the `startTable` property of the ingress descriptor

Example:

```console
$ lino id set-child-where film_original_language_id_fkey "creation_date >= '01/01/2023'"
successfully update relation film_original_language_id_fkey in ingress descriptor
```

### `--ingress-descriptor` argument

Ingress descriptor filename is parameterized with the `--ingress-descriptor` argument or its short alias `-i`.
this argument is present for all commands below.

### Display plan

The `display-plan` utilities explain the `lino`'s plan to extract data from database.

```bash
$ lino id display-plan
step 1 - extract rows from public.customer
step 2 - extract rows from public.store following →customer_store_id_fkey relationship for rows extracted at step 1, then follow →store_manager_staff_id_fkey →staff_store_id_fkey relationships (loop until data exhaustion)
step 3 - extract rows from public.address following →staff_address_id_fkey relationship for rows extracted at step 2
step 4 - extract rows from public.city following →address_city_id_fkey relationship for rows extracted at step 3
.
.
.
```

### Show graph

The `show-graph` create a graph of tables as node and relation as edge.

```bash
$ lino id customer show-graph
```

`lino` open your browser to visualize graph generated.

![Test Image 1](doc/img/lino-graph-export.svg)

## Pull

The `pull` sub-command create a **json** object for each line (jsonline format http://jsonlines.org/) of the first table.

```
$ lino pull source
{"active":1,"activebool":true,"address_id":5,"create_date":"2006-02-14T00:00:00Z","customer_address_id_fkey":{"address":"1913 Hanoi Way","address2":"","address_city_id_fkey":{"city":"Sasebo","city_country_id_fkey":{"country":"Japan","country_id":50,"last_update":"2006-02-15T09:44:00Z"},"city_id":463,"country_id":50,"last_update":"2006-02-15T09:45:25Z"},"address_id":5,"city_id":463,"district":"Nagasaki","last_update":"2006-02-15T09:45:30Z","phone":"28303384290","postal_code":"35200"},"customer_id":1,"customer_store_id_fkey":{"address_id":1,"last_update":"2006-02-15T09:57:12Z","manager_staff_id":1,"store_address_id_fkey":{"address":"47 MySakila Drive","address2":null,"address_city_id_fkey":{"city":"Lethbridge","city_country_id_fkey":{"country":"Canada","country_id":20,"last_update":"2006-02-15T09:44:00Z"},"city_id":300,"country_id":20,"last_update":"2006-02-15T09:45:25Z"},"address_id":1,"city_id":300,"district":"Alberta","last_update":"2006-02-15T09:45:30Z","phone":"","postal_code":""},"store_id":1,"store_manager_staff_id_fkey":{"active":true,"address_id":3,"email":"Mike.Hillyer@sakilastaff.com","first_name":"Mike","last_name":"Hillyer","last_update":"2006-05-16T16:13:11.79328Z","password":"8cb2237d0679ca88db6464eac60da96345513964","picture":"iVBORw0KWgo=","staff_address_id_fkey":{"address":"23 Workhaven Lane","address2":null,"address_city_id_fkey":{"city":"Lethbridge","city_country_id_fkey":{"country":"Canada","country_id":20,"last_update":"2006-02-15T09:44:00Z"},"city_id":300,"country_id":20,"last_update":"2006-02-15T09:45:25Z"},"address_id":3,"city_id":300,"district":"Alberta","last_update":"2006-02-15T09:45:30Z","phone":"14033335568","postal_code":""},"staff_id":1,"store_id":1,"username":"Mike"}},"email":"MARY.SMITH@sakilacustomer.org","first_name":"MARY","last_name":"SMITH","last_update":"2006-02-15T09:57:20Z","store_id":1}
```

### --filter-from-file argument

To sample a database from a given `id` list or ohter criteria `LINO` can read filters from a JSON Line file with the argument `--filter-from-file`.

Each line is a filter and `lino` apply it to the start table to extract data.

#### --filter

`--filter` argument overide filter's criteria from file.

#### --limit

`--limit` is applied for each line in filters file. With a filters's file of `N` lines and a limit of `L` lino could extract a maximum of `N` x `L` lines.

#### --where

`--where` argument is a raw SQL clause criteria (without `where` keyword) applied to the **start table only**. It's combined with `--filter` or `--filter-from-file` with the `and` operator.

#### --distinct

`--distinct` option (or `-D`) to return only distincts rows from the first table.

## Push

The `push` sub-command import a **json** line stream (jsonline format http://jsonlines.org/) in each table, following the ingress descriptor defined in current directory.

### How to update primary key

Let's say you have this record in database :

```json
{"actor_id":2,"last_name":"CHASE"}
```

And you want to change the primary key of this record to 998 :

```json
{"actor_id":998,"last_name":"CHASE"}
```

LINO will need the current key `2` to find the record in database, and the new key `998` to update the value. This can be done with the special field `__usingpk__` :

```json
{"actor_id":998,"last_name":"CHASE","__usingpk__":{"actor_id":2}}
```

The push update command will used the pk value inside `__usingpk__` to find the record, and the pk value from the actor object to update the record.

```console
lino push update source --table actor <<<'{"actor_id":998,"last_name":"CHASE","__usingpk__":{"actor_id":2}}'
```

The `__usingpk__` field can also be used with an ingress descriptor at any level in the data. The name of this field can be changed to another value with the `--using-pk-field` flag.

## Analyse

Use the `lino analyse <data_connector_alias>` command to extract metrics from the database in YAML format.

Only tables and columns explicitly listed in the tables.yaml file will be analysed.

Example result :

```yaml
database: source
tables:
    - name: first_name
      columns:
        - name: actor
          type: string
          concept: ""
          constraint: []
          confidential: null
          mainMetric:
            count: 200
            empty: 0
            unique: 128
            sample:
                - WALTER
                - MAE
                - LAURENCE
                - GREG
                - ALEC
          stringMetric:
            mostFrequentLen:
                - length: 4
                  freq: 0.235
                  sample:
                    - GARY
                    - ALAN
                    - ADAM
                    - JEFF
                    - GINA
                - length: 5
                  freq: 0.215
                  sample:
                    - REESE
                    - MILLA
                    - SALMA
                    - RALPH
                    - SUSAN
                - length: 7
                  freq: 0.16
                  sample:
                    - OLYMPIA
                    - KIRSTEN
                    - MATTHEW
                    - RICHARD
                    - KIRSTEN
                - length: 6
                  freq: 0.14
                  sample:
                    - WHOOPI
                    - WALTER
                    - SANDRA
                    - WHOOPI
                    - JOHNNY
                - length: 3
                  freq: 0.12
                  sample:
                    - BOB
                    - BEN
                    - KIM
                    - BOB
                    - TOM
            leastFrequentLen:
                - length: 11
                  freq: 0.01
                  sample:
                    - CHRISTOPHER
                - length: 9
                  freq: 0.02
                  sample:
                    - CHRISTIAN
                    - SYLVESTER
                - length: 2
                  freq: 0.02
                  sample:
                    - AL
                    - ED
                - length: 8
                  freq: 0.08
                  sample:
                    - JULIANNE
                    - LAURENCE
                    - JULIANNE
                    - SCARLETT
                    - LAURENCE
```

## Interaction with other tools

**LINO** respect the UNIX philosophy and use standards input an output to share data with others tools.

### MongoDB storage

Data set could be store in mongoDB easily with the `mongoimport` tool:

```
$ lino pull source --limit 100 | mongoimport --db myproject --collection customer
```

and reload later to a database :

```bash
$ mongoexport --db myproject --collection customer | lino push customer --jdbc jdbc:oracle:thin:scott/tiger@target:1721:xe
```

### Miller `mlr`

`mlr` tool can be used to format json lines into another tabular format (csv, markdown table, ...).

### jq

`jq` tool can be piped with the **LINO** output to prettify it.

```
$ lino pull source | jq
```

Pull sub field from the JSON stream

```
$ lino pull source --limit 3 | jq ".email"
"MARY.SMITH@sakilacustomer.org"
"PATRICIA.JOHNSON@sakilacustomer.org"
"LINDA.WILLIAMS@sakilacustomer.org"
```

Project subfield to produce other *JSON* objects

```
$ lino pull source --limit 3 | jq '{ "manager": .customer_store_id_fkey.store_manager_staff_id_fkey.first_name , "customer_email" :  .email }'

{
  "manager": "Mike",
  "customer_email": "MARY.SMITH@sakilacustomer.org"
}
{
  "manager": "Mike",
  "customer_email": "PATRICIA.JOHNSON@sakilacustomer.org"
}
{
  "manager": "Mike",
  "customer_email": "LINDA.WILLIAMS@sakilacustomer.org"
}
```

## Installation

Download the last binary release in your path.

## Contributors

* CGI France ✉[Contact support](mailto:LINO.fr@cgi.com)
* Pole Emploi

## License

Copyright (C) 2021 CGI France

LINO is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

LINO is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
 along with LINO.  If not, see <http://www.gnu.org/licenses/>.
