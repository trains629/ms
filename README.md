# 微服务架构

之前找到一个英汉词典库是用python写的，将其中读取词典的功能用能用go语言实现了一下，可以从词典库中读取词条。然后想着将这些词条批量的插入到postgresql数据库中，可以做成一个web应用，在插入的过程中遇到插入效率的问题，20万条词条依次插入数据库是相当耗费时间的，可以使用事物大数据指令等办法提高插入速度。如果不把这20万的数据看做词条而是不同用户输入数据，那么按单条插入是比较理想的模拟方式。现在准备将这个插入过程进行修改，采用微服务的构架，看是否能提供速度。

采用微服务的方式，就需要将整个功能进行分解，整个词典入口功能会产生平静的地方是数据读取，以及数据入库。在单一应用的过程中，词条是按顺序读取，然后利用go语言的并发机制压入数据库，这样就对数据库产生极大的压力，相当于一下子有几百上千的连接，会频繁的看到数据库无相应或超时。而改成一次一条又慢的厉害。所以现在准备在数据库前增加消息队列，读取的数据条先到消息队列，然后建立多个消息的消费者，这些消费者在将这些词条分别放入数据库，这样数据库的压力就分解到消息队列和多个消费者上。

这样整个功能需要包含以下服务：

1. 词条采集 读取词典中的词条，然后发送到消息队列中
2. 消息队列
3. 词条入库服务 将消息队列传送过来的词条数据插入数据库
4. postgresql 关系型数据库

采用微服务方式，就不太容易从控台查看状态，需要增加一个日志收集服务，将词条入库的过程进行汇总，也方便了解整个功能的效率。另外还需要增加配置中心，服务和数据库的信息都配置到配置中心，这样就需要一个服务发现的过程。然后还需要将整个系统简化，提供一个单独工具，用来创建数据表将消息队列等配置信息注册到配置中心（现在不知道怎么能将postgresql直接注册到配置中心）。

前面是整个功能的组成部分和要实现的功能。其中的服务发现，数据入库，消息队列等都是通用功能，准备建立一个公共函数库，这样除了第三方的服务还需要以下几个部分：

1. 公共函数库 提供日志记录，数据库操作，消息队列操作，服务注册
2. 初始化工具 初始化数据库，将数据库消息队列的信息注册到配置中心
3. 词条读取 从词典中读取词条并将词条发送给消息队列
4. 词条入库 将消息队列传递过来的词条写入数据库

为了统一管理和配置，将多数服务都配置在配置中心，按服务名称去访问配置中心的信息。另外也将消息队列的主题名称也配置在配置中心，服务启动之后将自己也注册到配置中心。

进行的是英文词典，消息队列的主题名称en-dict，频道为en。需要手动注册的服务有：

1. postgresql 目标数据库
2. elasticsearch 日志记录
3. nsq 消息队列

## 功能组成

### 公共函数库 base

包含公共操作的函数，诸如日志，数据库，消息队列等的操作和调用。

#### 日志

日志准备使用zerolog，然后通过hooks的方式将日志信息传递给elasticsearch，主要是没有找到zerolog是否可以直接将日志信息传递给elasticsearch。

需要将整个日志操作封装成对象，在服务启动的时候去执行初始化操作，然后返回的是zerolog的对象，使用这个对象去进行日志添加。在初始化的时候如果没有设置elasticsearch，那么就不需要将日志添加到elasticsearch中。

先将消息发送到消息队列，之前测试elasticsearch的时候一直没办法插入数据。

#### 数据库

封装一个postgresql的操作对象，通过数据库连接参数去建立连接，使用默认的数据库操作对象去执行数据库操作。

```sql
CREATE TABLE IF NOT EXISTS public.dict_en(
  id serial PRIMARY KEY,
  word text NOT NULL,
  pronunciation text[] NOT NULL,
  paraphrase text[] NOT NULL,
  rank text,
  pattern text,
  sentence jsonb,
  createtime timestamp without time zone NOT NULL DEFAULT LOCALTIMESTAMP,
  updatetime timestamp without time zone NOT NULL DEFAULT now()
)
```

#### 消息队列

消息队列准备采用nsq，分成两个部分，一个是消息的消费，一个是消息的生成

#### 配置中心

配置中心准备采用etcd，在这里按服务名称去添加查找服务的配置信息，在服务名称前增加统一的前缀`/dict`。

### 初始化工具 init

这个部分是独立的工具，负责向配置中心写入目标数据库，日志服务，消息队列地址。

将数据写入etcd的服务中，需要写入以下服务：

1. postgresql
2. nsq

这样准备先将信息写到yml的配置文件中，然后使用etcd接口读取。
按服务名称读取服务列表中的服务，然后将信息以json串形式写入etcd中。

之前使用的`google.golang.org/grpc v1.27.1 // indirect`，结果一直报错，这个项目是新创建的，之前在其他库中一直都是正常的，经过对比正常的库是`google.golang.org/grpc v1.25.1 // indirect`,在go.mod中将这个库替换为低版本就可以正常运行。

```
go run ./main.go                                                 
# github.com/coreos/etcd/clientv3/balancer/picker
../../go/pkg/mod/github.com/coreos/etcd@v3.3.18+incompatible/clientv3/balancer/picker/err.go:37:44: undefined: balancer.PickOptions
../../go/pkg/mod/github.com/coreos/etcd@v3.3.18+incompatible/clientv3/balancer/picker/roundrobin_balanced.go:55:54: undefined: balancer.PickOptions
# github.com/coreos/etcd/clientv3/balancer/resolver/endpoint
../../go/pkg/mod/github.com/coreos/etcd@v3.3.18+incompatible/clientv3/balancer/resolver/endpoint/endpoint.go:114:78: undefined: resolver.BuildOption
../../go/pkg/mod/github.com/coreos/etcd@v3.3.18+incompatible/clientv3/balancer/resolver/endpoint/endpoint.go:182:31: undefined: resolver.ResolveNowOption

```
 
### 词条读取 reader 

### 词条入库服务 writer

启动后先从etcd中读取数据库，消息队列信息，如果没有读取到etcd，重复执行3次检查，无效就结束当前应用。如果能读取到，分别去尝试连接消息队列和数据库，如果都不能连接也结束当前应用。正常之后应该将自己注册到etcd上，这样可以放便其他应用知道当前服务是否可用。

连接正常之后，检查是否创建数据表，如果没有创建先创建，然后一直监听目标主题，当有消息进来，执行入库操作。按nsq的数据流转方式，在一个频道下的消费者不会被发送相同的消息，这样只要词条读取不重复，这里应该也不会重复。

所有的数据都是通过etcd的配置中读取的，需要访问的服务名称都是固定，这样只需要传递etcd的服务地址。



## ServiceConfig

终于算是构建了一个基本的etcd读取对象，可以读取更新和读取etcd中的数据，接下来可以继续完成整个微服务框架的构建了。
