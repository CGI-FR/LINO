![GitHub 工作流状态](https://img.shields.io/github/actions/workflow/status/CGI-FR/LINO/ci.yml?branch=main)
[![Go 报告卡](https://goreportcard.com/badge/github.com/cgi-fr/pimo)](https://goreportcard.com/report/github.com/cgi-fr/lino)
![GitHub 所有发布](https://img.shields.io/github/downloads/CGI-FR/LINO/total)
![GitHub](https://img.shields.io/github/license/CGI-FR/LINO)
![GitHub 仓库星标](https://img.shields.io/github/stars/CGI-FR/LINO)
![GitHub go.mod Go 版本](https://img.shields.io/github/go-mod/go-version/CGI-FR/LINO)
![GitHub 最新发布日期](https://img.shields.io/github/v/release/CGI-FR/LINO)

[![en](https://img.shields.io/badge/lang-en-red.svg)](https://github.com/CGI-FR/LINO/blob/main/README.md)
[![zh-cn](https://img.shields.io/badge/lang-zh-cn-green.svg)](https://github.com/CGI-FR/LINO/blob/main/README.zh-cn.md)

# LINO : 大输入，窄输出

LINO 是一个简单的 ETL（提取转换加载）工具，用于管理测试数据。lino 命令行工具从关系型数据库中提取测试数据，以创建一个最接近生产环境的最小数据库。

![lino](doc/img/lino.png)

## 使用方法

`lino` 命令行工具在相对项目目录中工作，类似于  `git` 或 `docker`

可以使用以下标志：

* `--verbosity <level>` 或  `-v<level>` 此标志增加了 error日志 输出的详细程度，可选值为：none（0）、error（1）、warn（2）、info（3）、debug（4）、trace（5）。
* `--debug` 此标志用于在日志中添加调试信息。
* `--log-json` 设置此标志以生成 JSON 格式的日志，更详细地记录和结构化日志。
* `color` 此标志允许在控制台日志中添加颜色。
* `--stats <filename | url>` 此标志将运行统计信息输出到指定文件或将其发送到指定的 URL（必须以 `http` 或  `https`开头）。
* `--statsTemplate <string>` 此标志将使 LINO 使用该值作为模板来生成统计信息。请使用 Go 模板格式来包含统计信息。要包含它们，您必须将它们指定为 `{{ .Stats }}`. （例如 `{"software":"LINO","stats":{{ .Stats }}}`）。

## 数据库支持

### db2

Lino 支持 db2 系列的数据库，使用以下驱动程序： https://github.com/ibmdb/go_ibm_db.

### Mariadb / MySql

Lino 支持 MariaDB / MySql 数据库，使用以下驱动程序： <https://github.com/go-sql-driver/mysql>.

### Oracle DB

Lino 支持 Oracle 数据库，使用以下驱动程序： <https://github.com/sijms/go-ora>.

### Postgres

Lino 支持 Postgres 数据库，使用以下驱动程序：<https://github.com/lib/pq>.

### SQL Server

Lino 为 SQL Server 提供全面的支持，扩展兼容性包括 SQL Server、Azure SQL 数据库和 Azure SQL 实例管理。管理功能也与这些环境完全兼容。

通过 Microsoft 对 `ms-mssqldb` 驱动程序的调整实现了与 SQL Server 和 Go 编程语言之间的集成，这种兼容性的实现使得在 SQL Server 和 Go 环境之间进行更无缝、更高效的交互成为可能。您可以在 GitHub 上找到此驱动程序的链接： [https://github.com/microsoft/go-mssqldb](https://github.com/microsoft/go-mssqldb). 这一举措现在为开发人员提供了一个技术上优化的解决方案，以满足其数据管理和操作需求。

然而，由于在 SQL Server 上不可用 `TRUNCATE` 函数，它已被 `DELETE` 语句替换。因此，可以预期会有轻微的性能损失。

## 创建一个新的 LINO 项目

```
$ mkdir myproject
$ cd myproject
```

## 添加数据连接器

数据连接器是数据库连接的快捷方式。

```bash
$ lino dataconnector add source postgresql://postgres:sakila@localhost:5432/postgres?sslmode=disable
successfully added dataconnector {source postgresql://postgres:sakila@localhost:5432/postgres?sslmode=disable}
```

### 连接字符串

Lino 使用 URL 模式的连接字符串。

```
<databaseVendor>://[<user>[:<password>]]@<host>/<database>
```

目前支持的数据库类型有：

* postgresql
* oracle
* oracle-raw (用于完整的 TNS 支持 `oracle-raw://user:pwd@(DESCRIPTION=(ADDRESS=(PROTOCOL=TCP)(HOST=dbhost.example.com)(PORT=1521))(CONNECT_DATA=(SERVICE_NAME=orclpdb1)))`)
* mysql / mariadb
* db2 (alpha feature) : DB2 驱动程序目前正在开发中，请联系我们以获取带有 DB2 支持的 LINO 二进制文件的编译，以适配您的目标操作系统/架构
* http : 使用 HTTP 端点推送和拉取数据（适用于 golang 不支持的数据库）

### dataconnector.yml

生成的 `dataconnector.yml` 的内容是

```yaml
version: v1
dataconnectors:
  - name: source
    url: postgresql://postgres:sakila@localhost:5432/postgres?sslmode=disable
```

## 创建 relationships

LINO 创建一个一致的示例数据库。为了执行遵循外键约束的提取，LINO 必须提取表之间的关系。

使用 `relation` 子命令或其简称 `rel` 来从外键约束中提取关系。

```
$ lino relation extract source
lino finds 40 relations from constraints
```

生成的  `relations.yml` 的内容是

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

用户可以手动编辑 `relations.yml` 来添加不属于数据库模型的关系。

## 提取 Tables

`table` 动作提取有关数据库表的信息。

```
$ lino table extract source
lino finds 15 table(s)
```

`lino` 将表描述与列信息存储在 `table.yml` 文件中，列和键按照数据库的顺序排列：

```yaml
version: v1
tables:
  - name: actor
    keys:
      - actor_id
    columns:
      - name: actor_id
      - name: first_name
        export: string
      - name: last_name
        export: string
      - name: last_update
        export: timestamp
  - name: address
    keys:
```

`export:`  字段用于包含需要转换以正确显示的数据类型的列。

如果某些不寻常的类型无法自动检测到，附加 `-v2` 标志会启用日志信息。例如，您将收到类似于以下示例的通知，指示表的哪些列缺少类型：

`表 film 包含一些具有不寻常特征的列：[rating]。如果数据显示不正确，可能需要手动指定导出类型。`

### --only-tables

包括 `only-tables` 标志允许 LINO 仅提取表信息，排除列信息，并将其存储在 `table.yml` 文件中：

```yaml
version: v1
tables:
  - name: actor
    keys:
      - actor_id
  - name: address
    keys:
      - address_id
  - name: category
    keys:
      - category_id
  - name: city
    keys:
```

#### --with-db-infos

包括 `with-db-infos` 标志允许 LINO 提取更多关于列的信息，例如列类型、长度、大小和精度，如果列已配置了这些规格，则所有信息都将存储在`table.yml` 文件中：

```yaml
version: v1
tables:
  - name: actor
    keys:
      - actor_id
    columns:
      - name: actor_id
        dbinfo:
          type: INT4
      - name: first_name
        export: string
        dbinfo:
          type: VARCHAR
          length: 45
      - name: rental_rate
        export: numeric
        dbinfo:
          type: NUMERIC
          size: 2
          precision: 4
```

## Ingress 描述符

Ingress 描述符对象描述了 `lino` 如何通过关系提取测试数据。

### 创建 Ingress 描述符

要创建 Ingress 描述符，请使用以提取的起始表为参数的 `id` 子命令。

```bash
$ lino id create public.customer
successfully created ingress descriptor
```

`lino` 将新的 Ingress 描述符存储在 `ingress-descriptor.yml` 文件中：

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

通过编辑 `ingress-descriptor.yml` 文件或使用专用命令来自定义 Ingress 描述符的提取计划。

例如，此版本的  `ingress-descriptor.yml` 将过滤出在 2023 年 01 月 01 日之前创建的语言对象：

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

`where` 属性可以设置在 `child` 或  `parent` 对象上。当在父对象上使用时，如果父对象由 where 子句过滤，则提取的数据将包含一个空的父对象。当在子对象上使用时，结果子列表将被过滤（如果所有子项都被过滤，则列表将为空）。

要修改 `ingress-descriptor.yml`, 可以使用一些命令而不是直接编辑文件：

- `lino id set-child-lookup <relation name> <true or false>` : 修改子对象的  `lookup` 属性
- `lino id set-parent-lookup <relation name> <true or false>` : 修改父对象的 `lookup`属性
- `lino id set-child-where <relation name> <where clause>` : 修改子对象的 `where` 属性
- `lino id set-parent-where <relation name> <where clause>` : 修改父对象的 `where` 属性
- `lino id set-start-table <table name>` : 修改 Ingress 描述符的 `startTable` 属性

示例：

```console
$ lino id set-child-where film_original_language_id_fkey "creation_date >= '01/01/2023'"
successfully update relation film_original_language_id_fkey in ingress descriptor
```

### `--ingress-descriptor` 参数

Ingress 描述符文件名可以使用 `--ingress-descriptor` 参数或其简写别名 `-i` 进行参数化。此参数在以下所有命令中都存在。

### Display plan

`display-plan` 实用工具解释了 `lino` 从数据库中提取数据的计划。

```bash
$ lino id display-plan
step 1 - extract rows from public.customer (从 public.customer 中提取行)
step 2 - extract rows from public.store following →customer_store_id_fkey relationship for rows extracted at step 1, then follow →store_manager_staff_id_fkey →staff_store_id_fkey relationships (loop until data exhaustion)
(从 public.store 中提取行，按照 →customer_store_id_fkey 关系提取步骤 1 中提取的行，然后按照 →store_manager_staff_id_fkey →staff_store_id_fkey 关系提取（循环直至数据耗尽）)
step 3 - extract rows from public.address following →staff_address_id_fkey relationship for rows extracted at step 2
从 public.address 中提取行，按照 →staff_address_id_fkey 关系提取步骤 2 中提取的行
step 4 - extract rows from public.city following →address_city_id_fkey relationship for rows extracted at step 3
从 public.city 中提取行，按照 →address_city_id_fkey 关系提取步骤 3 中提取的行
.
.
.
```

### Show graph

`show-graph` 创建一个将表格作为节点、关系作为边的图形。

```bash
$ lino id customer show-graph
```

`lino` 打开您的浏览器以可视化生成的图形。

![Test Image 1](doc/img/lino-graph-export.svg)

## Pull

`pull` 子命令为第一个表的每一行创建一个 **json**  对象 (jsonline format http://jsonlines.org/).

```
$ lino pull source
{"active":1,"activebool":true,"address_id":5,"create_date":"2006-02-14T00:00:00Z","customer_address_id_fkey":{"address":"1913 Hanoi Way","address2":"","address_city_id_fkey":{"city":"Sasebo","city_country_id_fkey":{"country":"Japan","country_id":50,"last_update":"2006-02-15T09:44:00Z"},"city_id":463,"country_id":50,"last_update":"2006-02-15T09:45:25Z"},"address_id":5,"city_id":463,"district":"Nagasaki","last_update":"2006-02-15T09:45:30Z","phone":"28303384290","postal_code":"35200"},"customer_id":1,"customer_store_id_fkey":{"address_id":1,"last_update":"2006-02-15T09:57:12Z","manager_staff_id":1,"store_address_id_fkey":{"address":"47 MySakila Drive","address2":null,"address_city_id_fkey":{"city":"Lethbridge","city_country_id_fkey":{"country":"Canada","country_id":20,"last_update":"2006-02-15T09:44:00Z"},"city_id":300,"country_id":20,"last_update":"2006-02-15T09:45:25Z"},"address_id":1,"city_id":300,"district":"Alberta","last_update":"2006-02-15T09:45:30Z","phone":"","postal_code":""},"store_id":1,"store_manager_staff_id_fkey":{"active":true,"address_id":3,"email":"Mike.Hillyer@sakilastaff.com","first_name":"Mike","last_name":"Hillyer","last_update":"2006-05-16T16:13:11.79328Z","password":"8cb2237d0679ca88db6464eac60da96345513964","picture":"iVBORw0KWgo=","staff_address_id_fkey":{"address":"23 Workhaven Lane","address2":null,"address_city_id_fkey":{"city":"Lethbridge","city_country_id_fkey":{"country":"Canada","country_id":20,"last_update":"2006-02-15T09:44:00Z"},"city_id":300,"country_id":20,"last_update":"2006-02-15T09:45:25Z"},"address_id":3,"city_id":300,"district":"Alberta","last_update":"2006-02-15T09:45:30Z","phone":"14033335568","postal_code":""},"staff_id":1,"store_id":1,"username":"Mike"}},"email":"MARY.SMITH@sakilacustomer.org","first_name":"MARY","last_name":"SMITH","last_update":"2006-02-15T09:57:20Z","store_id":1}
```

### --filter-from-file 参数

要从给定的 `id` 列表或其他条件对数据库进行采样，`LINO` 可以从 JSON Line 文件中读取过滤器，使用 `--filter-from-file` 参数。

每行都是一个过滤器，lino 将其应用于起始表以提取数据。

#### --filter

`--filter` 参数覆盖文件中的过滤器条件。

#### --limit

`--limit` 应用于文件中的每行过滤器。对于具有 `N` 行和限制 `L` 的过滤器文件，lino 可以最多提取 `N` x `L` 行。

#### --where

`--where` 参数是应用于 **仅起始表** 的原始 SQL 子句条件（不带 `where` 关键字）。它与 `--filter` 或 `--filter-from-file` 结合使用，使用 `and` 运算符。

#### --distinct

`--distinct` 选项(或 `-D`) 只返回第一个表中的不同行。

## Push

The `push` 子命令将 **json** 行流 (jsonline 格式 http://jsonlines.org/) 导入到每个表中，按照当前目录中定义的 `Ingress` 描述符。

### 自动截断值

使用 `autotruncate` 标志自动截断超出数据库接受的最大长度的字符串值。

```
$ lino push truncate dest --table actor --autotruncate < actors.jsonl
```

LINO 将根据表格.yaml文件中每个列设置的每个 `dbinfo`.`length` 参数来截断每个值。

此外，如果您的数据库最大值不是以字符数而是以字节数定义的，请将 `dbinfo`.`bytes` 设置为 true。LINO 将根据最大字节数而不是字符数（目前假设使用 utf-8 编码）来截断该值。

```yaml
version: v1
tables:
  - name: actor
    keys:
      - actor_id
    columns:
      - name: actor_id
        dbinfo:
          type: INT4
      - name: first_name
        export: string
        dbinfo:
          type: VARCHAR
          length: 45
          bytes: true
```

### 如何更新主键

假设您在数据库中有以下记录：

```json
{"actor_id":2,"last_name":"CHASE"}
```

并且您希望将此记录的主键更改为 998：

```json
{"actor_id":998,"last_name":"CHASE"}
```

LINO 需要当前键 `2` 来在数据库中找到记录，并且需要新键 `998` 来更新该值。这可以通过特殊字段 `__usingpk__` 来完成：

```json
{"actor_id":998,"last_name":"CHASE","__usingpk__":{"actor_id":2}}
```


如何更新主键
假设您在数据库中有以下记录：

json
Copy code
{"actor_id":2,"last_name":"CHASE"}
并且您希望将此记录的主键更改为 998：

json
Copy code
{"actor_id":998,"last_name":"CHASE"}
LINO 需要当前键 2 来在数据库中找到记录，并且需要新键 998 来更新该值。这可以通过特殊字段 __usingpk__ 来完成：

json
Copy code
{"actor_id":998,"last_name":"CHASE","__usingpk__":{"actor_id":2}}
推送更新命令将使用 `__usingpk__` 中的主键值来查找记录，并使用来自 actor 对象的主键值来更新记录。

```console
lino push update source --table actor <<<'{"actor_id":998,"last_name":"CHASE","__usingpk__":{"actor_id":2}}'
```

`__usingpk__` 字段还可以与数据中的 Ingress 描述符一起使用，在任何级别。此字段的名称可以使用 `--using-pk-field` 标志更改为其他值。

### 如何从错误中恢复

结合使用选项 `lino pull --exclude-from-file` (简写为 `-X`) 和 `lino push --savepoint`  可以处理错误恢复。如果错误中断了先前运行的过程，则该过程将从失败的位置重新启动。

示例：

```console
lino pull --table customers -X customers.jsonl source | lino push --table customers --savepoint customers.jsonl target
```

customers.jsonl 文件将包含已传输到目标数据库的客户 ID 列表。

## Analyse

使用 `lino analyse <data_connector_alias>`命令从数据库中提取指标，并以 YAML 格式输出。

只有在 tables.yaml 文件中明确列出的表和列才会被分析。

示例结果：

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

## 与其他工具的交互

**LINO** 遵循 UNIX 哲学，使用标准输入和输出与其他工具共享数据。

### MongoDB 存储

可以使用 `mongoimport` 工具轻松地将数据集存储在 MongoDB 中：

```
$ lino pull source --limit 100 | mongoimport --db myproject --collection customer
```

然后可以从数据库重新加载数据：

```bash
$ mongoexport --db myproject --collection customer | lino push customer --jdbc jdbc:oracle:thin:scott/tiger@target:1721:xe
```

### Miller `mlr`

可以使用 `mlr` 工具将 json 行格式化为其他表格格式（csv、markdown 表格等）。

### jq

可以将 `jq`工具与 **LINO** 输出管道连接以美化它。

```
$ lino pull source | jq
```

从 JSON 流中提取子字段

```
$ lino pull source --limit 3 | jq ".email"
"MARY.SMITH@sakilacustomer.org"
"PATRICIA.JOHNSON@sakilacustomer.org"
"LINDA.WILLIAMS@sakilacustomer.org"
```

投影子字段以生成其他 JSON 对象

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

## 安装

在您的路径中下载最新的二进制发布版。

## 贡献者

* CGI France ✉[联系支持](mailto:LINO.fr@cgi.com)
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

版权所有 (C) 2021 CGI France

LINO 是自由软件：您可以重新分发和/或修改
它遵循 GNU 通用公共许可证发布的条款
自由软件基金会，版本 3 的许可证，或
（按您的选择）任何更高版本。

LINO 是根据以下条款分发的希望它有用，
但是没有任何担保；甚至没有暗示的担保
适销性或适用于特定用途。有关更多详细信息，请参见
GNU 通用公共许可证。

您应该已经收到了 GNU 通用公共许可证的副本
连同 LINO 一起。如果没有，请参阅 http://www.gnu.org/licenses/。
