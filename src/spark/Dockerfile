FROM bitnami/spark:3.5.1-debian-11-r0

USER root
WORKDIR /app
RUN wget -P /opt/bitnami/spark/jars https://repo1.maven.org/maven2/org/apache/spark/spark-sql-kafka-0-10_2.12/3.5.1/spark-sql-kafka-0-10_2.12-3.5.1.jar \
    && wget -P /opt/bitnami/spark/jars https://repo1.maven.org/maven2/org/apache/kafka/kafka-clients/3.6.0/kafka-clients-3.6.0.jar