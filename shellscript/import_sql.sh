#!/usr/bin/env bash
set -e

sql_path=$1
if [ -z "$sql_path" ]; then
  echo "Usage: import_sql.sh <sql_path>"
  exit 1
fi

assert_cmd_exists() {
  if type $1 >/dev/null 2>&1; then
      return
  else
      echo "Fatal: $1 command does not exist, exited."
      exit 1
  fi
}

assert_cmd_exists mysql

# 数据库配置
DB_HOST=127.0.0.1 # Cannot be localhost!
DB_USER=root
DB_PASSWORD="123"


# 获取当前目录下的所有 .sql 文件
sql_files=$(ls $sql_path/*.sql)

# 导入每个 SQL 文件
for file in $sql_files
do
    echo "Importing $file..."
    mysql -h $DB_HOST -u $DB_USER -p$DB_PASSWORD < "$file"
done

echo "done."




