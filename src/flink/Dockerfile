FROM maven:3-eclipse-temurin-21 AS build

COPY pom.xml .
RUN mvn dependency:go-offline

COPY . .
RUN mvn clean verify -o

FROM flink:1.20

RUN cd /opt/flink/lib && \
   curl -sO https://repo1.maven.org/maven2/org/apache/logging/log4j/log4j-layout-template-json/2.17.1/log4j-layout-template-json-2.17.1.jar
   
COPY --from=build --chown=flink:flink target/candle-stick-job-1.0.jar /opt/flink-jobs/candle-stick-job-1.0.jar


RUN echo "execution.checkpointing.interval: 10s" >> /opt/flink/conf/flink-conf.yaml; \
    echo "pipeline.object-reuse: true" >> /opt/flink/conf/flink-conf.yaml; \
    echo "pipeline.time-characteristic: EventTime" >> /opt/flink/conf/flink-conf.yaml; \
    echo "taskmanager.memory.jvm-metaspace.size: 256m" >> /opt/flink/conf/flink-conf.yaml; \
    echo "parallelism.default: 1" >> /opt/flink/conf/flink-conf.yaml;