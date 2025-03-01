FROM maven:3-eclipse-temurin-21 AS build

COPY pom.xml .
RUN mvn dependency:go-offline

COPY . .
RUN mvn clean verify -o

FROM flink:1.20-java17

RUN cd /opt/flink/lib && \
   curl -sO https://repo1.maven.org/maven2/org/apache/logging/log4j/log4j-layout-template-json/2.17.1/log4j-layout-template-json-2.17.1.jar
   
COPY --from=build --chown=flink:flink target/hello-world-job-1.0.jar /opt/flink-jobs/hello-world-job-1.0.jar