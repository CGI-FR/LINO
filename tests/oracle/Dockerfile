FROM store/oracle/database-enterprise:12.2.0.1

COPY assets /
RUN sqlplus sys/Oradoc_db1@//localhost:1521/ORCLCDB oracle-sakila-schema.sql