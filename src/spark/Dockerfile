FROM bitnami/spark:3.5.5-debian-12-r2

USER root
RUN apt-get update && apt-get install -y wget \
    && rm -rf /var/lib/apt/lists/* \
    && wget -P /opt/bitnami/spark/jars https://repo1.maven.org/maven2/org/apache/spark/spark-sql-kafka-0-10_2.12/3.5.5/spark-sql-kafka-0-10_2.12-3.5.5.jar \
    && wget -P /opt/bitnami/spark/jars https://repo1.maven.org/maven2/org/apache/kafka/kafka-clients/3.6.0/kafka-clients-3.6.0.jar
USER 1001  # Retour à l’utilisateur par défaut de Bitnami
WORKDIR /app