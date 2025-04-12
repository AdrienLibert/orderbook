FROM python:3.12.1-bullseye

RUN apt-get update && apt-get install -y openjdk-11-jdk \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

ENV JAVA_HOME=/usr/lib/jvm/java-11-openjdk-amd64
ENV PATH=$PATH:$JAVA_HOME/bin

# Install Spark
RUN curl -SL https://archive.apache.org/dist/spark/spark-3.5.5/spark-3.5.5-bin-hadoop3.tgz -o spark-3.5.5-bin-hadoop3.tgz \
    && tar -xzf spark-3.5.5-bin-hadoop3.tgz -C /opt/ \
    && ln -s /opt/spark-3.5.5-bin-hadoop3 /opt/spark \
    && rm spark-3.5.5-bin-hadoop3.tgz

ENV SPARK_HOME=/opt/spark
ENV PATH=$PATH:$SPARK_HOME/bin

RUN pip install --no-cache-dir pyspark==3.5.5

RUN mkdir /app

RUN curl -SL https://jdbc.postgresql.org/download/postgresql-42.7.3.jar -o /app/postgresql-42.7.3.jar
RUN curl -SL https://repo1.maven.org/maven2/org/apache/spark/spark-sql-kafka-0-10_2.12/3.5.5/spark-sql-kafka-0-10_2.12-3.5.5.jar -o /app/spark-sql-kafka-0-10_2.12-3.5.5.jar
RUN curl -SL https://repo1.maven.org/maven2/org/apache/kafka/kafka-clients/3.5.2/kafka-clients-3.5.2.jar -o /app/kafka-clients-3.5.2.jar
RUN curl -SL https://repo1.maven.org/maven2/org/apache/spark/spark-token-provider-kafka-0-10_2.12/3.5.5/spark-token-provider-kafka-0-10_2.12-3.5.5.jar -o /app/spark-token-provider-kafka-0-10_2.12-3.5.5.jar

COPY ./spark/kafka_spark.py /app/

WORKDIR /app

CMD ["/opt/spark/bin/spark-submit", "--jars", "/app/spark-sql-kafka-0-10_2.12-3.5.5.jar,/app/kafka-clients-3.5.2.jar,/app/spark-token-provider-kafka-0-10_2.12-3.5.5.jar", "/app/kafka_spark.py"]