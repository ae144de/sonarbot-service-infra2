# infra/create-topics.sh
#!/bin/bash
set -e

KAFKA=bash /usr/bin/kafka-topics

# Eğer yoksa yarat
/usr/bin/kafka-topics --bootstrap-server kafka:9092 \
  --create --if-not-exists --topic kline.raw \
  --partitions 1 --replication-factor 1

/usr/bin/kafka-topics --bootstrap-server kafka:9092 \
  --create --if-not-exists --topic analysis.request \
  --partitions 1 --replication-factor 1

/usr/bin/kafka-topics --bootstrap-server kafka:9092 \
  --create --if-not-exists --topic alert.trigger \
  --partitions 1 --replication-factor 1

/usr/bin/kafka-topics --bootstrap-server kafka:9092 \
  --create --if-not-exists --topic indicator.calc \
  --partitions 1 --replication-factor 1

/usr/bin/kafka-topics --bootstrap-server kafka:9092 \
  --create --if-not-exists --topic news.incoming \
  --partitions 1 --replication-factor 1

/usr/bin/kafka-topics --bootstrap-server kafka:9092 \
  --create --if-not-exists --topic trade.exec \
  --partitions 1 --replication-factor 1

/usr/bin/kafka-topics --bootstrap-server kafka:9092 \
  --create --if-not-exists --topic test.request \
  --partitions 1 --replication-factor 1