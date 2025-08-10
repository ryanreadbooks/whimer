FROM bitnami/elasticsearch:8.18.0

RUN /opt/bitnami/elasticsearch/bin/elasticsearch-plugin install -b https://get.infini.cloud/elasticsearch/analysis-ik/8.18.0
