FROM gradle:8.14.3 AS build
WORKDIR /app
COPY . /app
RUN  chmod +x ./gradlew
RUN ./gradlew build -x test


FROM eclipse-temurin:21-jdk AS production
COPY --from=build /app/build/libs/FreStyle-0.0.1-SNAPSHOT.jar ./FreStyle.jar
CMD [ "java", "-jar", "FreStyle.jar" ]