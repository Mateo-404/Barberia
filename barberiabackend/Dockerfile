# 1️⃣ Build con Maven
FROM maven:3.9.0-eclipse-temurin-21 AS build
WORKDIR /app

# Copiar pom y código fuente
COPY pom.xml .
COPY src ./src

# Build del jar ejecutable, omitiendo tests
RUN mvn clean package -DskipTests

# 2️⃣ Imagen liviana para correr la app
FROM eclipse-temurin:21-jdk-jammy
WORKDIR /app

# Copiar el jar generado
COPY --from=build /app/target/barberBackend-0.0.1-SNAPSHOT.jar app.jar

# Exponer el puerto que usa Spring Boot
EXPOSE 8080

# Comando para correr la app
ENTRYPOINT ["java","-jar","app.jar"]
