FROM python:3.12.1-bullseye

RUN apt-get update && apt-get install -y openjdk-11-jdk

ENV JAVA_HOME=/usr/lib/jvm/java-11-openjdk-amd64

RUN pip install --no-cache-dir pyspark==3.5.5

RUN mkdir /app
RUN wget -P 3.5.5/jars/ https://jdbc.postgresql.org/download/postgresql-42.7.4.jar;


COPY ./spark /app/

WORKDIR /app

ENTRYPOINT ["python3", "trades_spark.py"]