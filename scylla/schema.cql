CREATE KEYSPACE IF NOT EXISTS censys WITH replication = {
  'class': 'SimpleStrategy',
  'replication_factor': 1
};
CREATE TABLE IF NOT EXISTS censys.services (
  ip inet,
  port int,
  service text,
  data text,
  timestamp timestamp,
  PRIMARY KEY (ip, port)
);
