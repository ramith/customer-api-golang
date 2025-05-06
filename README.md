# customer-api-golang

A sample microservice written in Go for demo purposes.

This service connects to a MySQL database and is deployed on [WSO2 Choreo](https://wso2.com/choreo/).

## Database Setup

This service requires a MySQL database named `customerdb`.

Run the following DDL to create the required `customer` table:

```sql
CREATE TABLE `customer` (
  `account_id` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci NOT NULL,
  `first_name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL,
  `last_name` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL,
  `kyc_status` varchar(100) CHARACTER SET utf8mb4 COLLATE utf8mb4_0900_ai_ci DEFAULT NULL,
  PRIMARY KEY (`account_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_0900_ai_ci;

```
# Required Environment Variables
Set the following environment variables to enable the database connection:

 **MYSQL_HOST** – MySQL server hostname (e.g., localhost)

 **MYSQL_USER** – MySQL username

 **MYSQL_PWD** – MySQL user password


