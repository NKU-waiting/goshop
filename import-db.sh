#!/bin/bash

SQL_FILE="/home/waiting/goshop/newbee-mall-api-go/static-files/newbee_mall_db_v2_schema.sql"

echo "Importing database..."

# 导入 SQL 文件，跳过有问题的行
docker-compose exec -T mysql mysql -uroot -proot newbee_mall_db_v2 << 'EOF'
SET NAMES utf8mb4;
SET FOREIGN_KEY_CHECKS = 0;
EOF

# 读取 SQL 文件并跳过有问题的行
grep -v "SET SQL_NOTES" "$SQL_FILE" | grep -v "SET SQL_MODE" | grep -v "SET FOREIGN_KEY_CHECKS" | grep -v "SET CHARACTER_SET_CLIENT" | grep -v "SET CHARACTER_SET_RESULTS" | grep -v "SET COLLATION_CONNECTION" | docker-compose exec -T mysql mysql -uroot -proot newbee_mall_db_v2

echo "Database import complete!"
